package smtp

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/textproto"
	"strings"
	"testing"
	"time"
)

func TestSendMailUnsafeUsesFinalDataResponseAsDeliveryResult(t *testing.T) {
	tests := []struct {
		name                 string
		dataResponse         string
		holdWithoutQuitReply bool
		wantCode             int
		maxDuration          time.Duration
	}{
		{
			name:                 "accepted message returns while server withholds QUIT reply",
			dataResponse:         "250 2.0.0 queued",
			holdWithoutQuitReply: true,
			maxDuration:          time.Second,
		},
		{
			name:         "DATA rejection remains a delivery failure",
			dataResponse: "550 5.0.0 rejected",
			wantCode:     550,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, serverDone := startSMTPTransactionServer(t, tt.dataResponse, tt.holdWithoutQuitReply)

			start := time.Now()
			err := SendMailUnsafe(
				"",
				addr,
				nil,
				"sender@example.com",
				"example.com",
				[]string{"recipient@example.net"},
				[]byte("From: sender@example.com\r\nTo: recipient@example.net\r\nSubject: test\r\n\r\nbody\r\n"),
			)
			duration := time.Since(start)
			if tt.maxDuration > 0 && duration > tt.maxDuration {
				t.Fatalf("SendMailUnsafe() took %v, want at most %v", duration, tt.maxDuration)
			}

			if tt.wantCode == 0 {
				if err != nil {
					t.Fatalf("SendMailUnsafe() = %v, want nil", err)
				}
			} else {
				var protocolErr *textproto.Error
				if !errors.As(err, &protocolErr) {
					t.Fatalf("SendMailUnsafe() error type = %T, want *textproto.Error", err)
				}
				if protocolErr.Code != tt.wantCode {
					t.Fatalf("SMTP code = %d, want %d", protocolErr.Code, tt.wantCode)
				}
			}

			select {
			case serverErr := <-serverDone:
				if serverErr != nil {
					t.Fatal(serverErr)
				}
			case <-time.After(2 * time.Second):
				t.Fatal("fake SMTP server did not finish")
			}
		})
	}
}

// withShortTimeouts 临时调小各阶段超时用于测试，返回恢复函数。
func withShortTimeouts(t *testing.T, greeting, command, dataStart, dataBlock, dataEnd time.Duration) {
	t.Helper()
	origGreeting, origCommand, origDataStart, origDataBlock, origDataEnd :=
		outboundSMTPGreetingTimeout, outboundSMTPCommandTimeout, outboundSMTPDataStartTimeout,
		outboundSMTPDataBlockTimeout, outboundSMTPDataEndTimeout
	outboundSMTPGreetingTimeout = greeting
	outboundSMTPCommandTimeout = command
	outboundSMTPDataStartTimeout = dataStart
	outboundSMTPDataBlockTimeout = dataBlock
	outboundSMTPDataEndTimeout = dataEnd
	t.Cleanup(func() {
		outboundSMTPGreetingTimeout, outboundSMTPCommandTimeout, outboundSMTPDataStartTimeout,
			outboundSMTPDataBlockTimeout, outboundSMTPDataEndTimeout =
			origGreeting, origCommand, origDataStart, origDataBlock, origDataEnd
	})
}

// 静默服务器（只建连不发 220）必须在问候超时内失败，
// 对应 RFC 5321 §4.5.3.2.1。
func TestNewClientBoundsSilentServer(t *testing.T) {
	withShortTimeouts(t, 50*time.Millisecond, 5*time.Minute, 2*time.Minute, 3*time.Minute, 10*time.Minute)

	clientConn, serverConn := net.Pipe()
	defer serverConn.Close()

	start := time.Now()
	client, err := NewClient(clientConn, "silent.example", "example.com")
	duration := time.Since(start)
	if client != nil {
		client.Close()
		t.Fatal("NewClient() returned a client from a silent server")
	}
	var netErr net.Error
	if !errors.As(err, &netErr) || !netErr.Timeout() {
		t.Fatalf("NewClient() error = %v, want timeout", err)
	}
	if duration > time.Second {
		t.Fatalf("NewClient() took %v, want at most 1s", duration)
	}
}

// 模拟一个"慢但健康"的下游：每个响应前都延迟一段时间。
// 总耗时必然超过任何单阶段超时，但只要按命令/数据块分别计时（而非
// 整个会话一个 Deadline），投递就应当成功。
func startSlowSMTPServer(t *testing.T, phaseDelay time.Duration) string {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() failed: %v", err)
	}

	go func() {
		defer listener.Close()
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		text := textproto.NewConn(conn)
		defer text.Close()

		steps := []struct {
			expectPrefix string
			response     string
		}{
			{"", "220 slow ESMTP ready"},
			{"EHLO ", "250 slow"},
			{"MAIL FROM:", "250 2.1.0 sender accepted"},
			{"RCPT TO:", "250 2.1.5 recipient accepted"},
			{"DATA", "354 end data with <CR><LF>.<CR><LF>"},
		}
		for _, step := range steps {
			time.Sleep(phaseDelay)
			if step.expectPrefix != "" {
				if err := expectSMTPCommand(text, step.expectPrefix); err != nil {
					return
				}
			}
			if err := text.PrintfLine("%s", step.response); err != nil {
				return
			}
		}

		if _, err := io.ReadAll(text.DotReader()); err != nil {
			return
		}
		time.Sleep(phaseDelay)
		if err := text.PrintfLine("250 2.0.0 queued"); err != nil {
			return
		}
		// 与已有用例一致：消息被接受后不等待 221。
		buf := make([]byte, 1)
		_, _ = conn.Read(buf)
	}()

	return listener.Addr().String()
}

// 各阶段分别计时后，一个每步都慢、但总时长超过单阶段超时的健康下游
// 不应被误杀；这正是"整个会话统一 Deadline"会出问题的场景。
func TestPhasedTimeoutsAllowSlowMultiPhaseTransaction(t *testing.T) {
	const phaseDelay = 150 * time.Millisecond
	// 每个阶段给足 3 倍余量，但单阶段超时（450ms）远小于总耗时（约 6*150ms=900ms）。
	withShortTimeouts(t, 5*time.Second, 450*time.Millisecond, 450*time.Millisecond, 450*time.Millisecond, 450*time.Millisecond)

	addr := startSlowSMTPServer(t, phaseDelay)

	err := SendMailUnsafe(
		"",
		addr,
		nil,
		"sender@example.com",
		"example.com",
		[]string{"recipient@example.net"},
		[]byte("From: sender@example.com\r\nTo: recipient@example.net\r\nSubject: slow\r\n\r\nbody\r\n"),
	)
	if err != nil {
		t.Fatalf("SendMailUnsafe() = %v, want nil (slow but healthy server must not be killed)", err)
	}
}

// 模拟下游在 354 之后停止读取数据：上传必须在单个数据块超时内失败，
// 而不是永久阻塞。
func startStalledDataServer(t *testing.T) string {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() failed: %v", err)
	}

	go func() {
		defer listener.Close()
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		text := textproto.NewConn(conn)
		defer text.Close()

		if err := text.PrintfLine("220 stalled ESMTP ready"); err != nil {
			return
		}
		if err := expectSMTPCommand(text, "EHLO "); err != nil {
			return
		}
		if err := text.PrintfLine("250 stalled"); err != nil {
			return
		}
		if err := expectSMTPCommand(text, "MAIL FROM:"); err != nil {
			return
		}
		if err := text.PrintfLine("250 2.1.0 sender accepted"); err != nil {
			return
		}
		if err := expectSMTPCommand(text, "RCPT TO:"); err != nil {
			return
		}
		if err := text.PrintfLine("250 2.1.5 recipient accepted"); err != nil {
			return
		}
		if err := expectSMTPCommand(text, "DATA"); err != nil {
			return
		}
		if err := text.PrintfLine("354 end data with <CR><LF>.<CR><LF>"); err != nil {
			return
		}

		// 关键：354 之后停止读取，模拟下游挂死。
		time.Sleep(30 * time.Second)
	}()

	return listener.Addr().String()
}

func TestStalledDataUploadFailsWithinDataBlockTimeout(t *testing.T) {
	withShortTimeouts(t, 5*time.Second, time.Minute, time.Minute, 300*time.Millisecond, time.Minute)

	addr := startStalledDataServer(t)

	// 构造 1MB 正文，足以填满内核 TCP 发送缓冲区，
	// 使写操作在下游停止读取后真正阻塞。
	body := make([]byte, 0, 1<<20)
	body = append(body, "From: sender@example.com\r\nTo: recipient@example.net\r\nSubject: stalled\r\n\r\n"...)
	for len(body) < 1<<20 {
		body = append(body, "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef\r\n"...)
	}

	start := time.Now()
	err := SendMailUnsafe("", addr, nil, "sender@example.com", "example.com", []string{"recipient@example.net"}, body)
	duration := time.Since(start)

	if err == nil {
		t.Fatal("SendMailUnsafe() = nil, want timeout error against stalled server")
	}
	var netErr net.Error
	if !errors.As(err, &netErr) || !netErr.Timeout() {
		t.Fatalf("SendMailUnsafe() error = %v, want timeout", err)
	}
	if duration > 10*time.Second {
		t.Fatalf("SendMailUnsafe() took %v, want failure within data-block timeout", duration)
	}
}

func startSMTPTransactionServer(t *testing.T, dataResponse string, holdWithoutQuitReply bool) (string, <-chan error) {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() failed: %v", err)
	}

	done := make(chan error, 1)
	go func() {
		defer listener.Close()

		conn, err := listener.Accept()
		if err != nil {
			done <- err
			return
		}
		defer conn.Close()
		_ = conn.SetDeadline(time.Now().Add(2 * time.Second))

		text := textproto.NewConn(conn)
		defer text.Close()

		if err := text.PrintfLine("220 test ESMTP ready"); err != nil {
			done <- err
			return
		}
		if err := expectSMTPCommand(text, "EHLO "); err != nil {
			done <- err
			return
		}
		if err := text.PrintfLine("250 test"); err != nil {
			done <- err
			return
		}
		if err := expectSMTPCommand(text, "MAIL FROM:"); err != nil {
			done <- err
			return
		}
		if err := text.PrintfLine("250 2.1.0 sender accepted"); err != nil {
			done <- err
			return
		}
		if err := expectSMTPCommand(text, "RCPT TO:"); err != nil {
			done <- err
			return
		}
		if err := text.PrintfLine("250 2.1.5 recipient accepted"); err != nil {
			done <- err
			return
		}
		if err := expectSMTPCommand(text, "DATA"); err != nil {
			done <- err
			return
		}
		if err := text.PrintfLine("354 end data with <CR><LF>.<CR><LF>"); err != nil {
			done <- err
			return
		}
		body, err := io.ReadAll(text.DotReader())
		if err != nil {
			done <- err
			return
		}
		if !strings.Contains(string(body), "Subject: test") {
			done <- fmt.Errorf("message body was not received: %q", body)
			return
		}
		if err := text.PrintfLine("%s", dataResponse); err != nil {
			done <- err
			return
		}

		if holdWithoutQuitReply {
			if err := expectSMTPCommand(text, "QUIT"); err != nil {
				done <- err
				return
			}

			// Keep the server side open without sending 221. The client must close
			// the accepted transaction without waiting for a reply.
			buf := make([]byte, 1)
			if _, err := conn.Read(buf); err != io.EOF {
				done <- fmt.Errorf("waiting for client close after QUIT: %w", err)
				return
			}
		}
		done <- nil
	}()

	return listener.Addr().String(), done
}

func expectSMTPCommand(conn *textproto.Conn, prefix string) error {
	line, err := conn.ReadLine()
	if err != nil {
		return err
	}
	if !strings.HasPrefix(line, prefix) {
		return fmt.Errorf("SMTP command = %q, want prefix %q", line, prefix)
	}
	return nil
}
