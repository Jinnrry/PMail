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
