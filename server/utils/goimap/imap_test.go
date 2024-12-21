package goimap

import (
	"net"
	"net/netip"
	"reflect"
	"testing"
	"time"
)

func Test_paramsErr(t *testing.T) {

}

func Test_getCommand(t *testing.T) {
	type args struct {
		line string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 string
		want2 string
	}{
		{
			"STATUS命令测试",
			args{`15.64 STATUS "Deleted Messages" (MESSAGES UIDNEXT UIDVALIDITY UNSEEN)`},
			"15.64",
			"STATUS",
			`"Deleted Messages" (MESSAGES UIDNEXT UIDVALIDITY UNSEEN)`,
		},
		{
			"LOGIN命令测试",
			args{`a LOGIN admin 666666`},
			"a",
			"LOGIN",
			`admin 666666`,
		},
		{
			"SELECT命令测试",
			args{`9.79 SELECT INBOX`},
			"9.79",
			"SELECT",
			`INBOX`,
		},
		{
			"CAPABILITY命令测试",
			args{`1.81 CAPABILITY`},
			"1.81",
			"CAPABILITY",
			``,
		},
		{
			"DELETE命令测试",
			args{`3.183 SELECT "Deleted Messages"`},
			"3.183",
			"SELECT",
			`"Deleted Messages"`,
		},
		{
			"异常命令测试",
			args{`GET/HTTP/1.0`},
			"",
			"",
			``,
		},
		{
			"FETCH命令测试",
			args{`4.189 FETCH 7:38 (INTERNALDATE UID RFC822.SIZE FLAGS BODY.PEEK[HEADER.FIELDS (date subject from to cc message-id in-reply-to references content-type x-priority x-uniform-type-identifier x-universally-unique-identifier list-id list-unsubscribe bimi-indicator bimi-location x-bimi-indicator-hash authentication-results dkim-signature)])`},
			"4.189",
			"FETCH",
			`7:38 (INTERNALDATE UID RFC822.SIZE FLAGS BODY.PEEK[HEADER.FIELDS (date subject from to cc message-id in-reply-to references content-type x-priority x-uniform-type-identifier x-universally-unique-identifier list-id list-unsubscribe bimi-indicator bimi-location x-bimi-indicator-hash authentication-results dkim-signature)])`,
		},
		{
			"FETCH命令测试2",
			args{`4.167 FETCH 1:41 (FLAGS UID)`},
			"4.167",
			"FETCH",
			`1:41 (FLAGS UID)`,
		},
		{
			"UID FETCH命令测试",
			args{`4.200 UID FETCH 5 BODY.PEEK[HEADER]`},
			"4.200",
			"UID FETCH",
			`5 BODY.PEEK[HEADER]`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2 := getCommand(tt.args.line)
			if got != tt.want {
				t.Errorf("getCommand() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("getCommand() got1 = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("getCommand() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

type mockConn struct{}

func (m mockConn) Read(b []byte) (n int, err error) {
	return 0, err
}

func (m mockConn) Write(b []byte) (n int, err error) {
	return 0, err
}

func (m mockConn) Close() error {
	return nil
}

func (m mockConn) LocalAddr() net.Addr {
	return net.TCPAddrFromAddrPort(netip.AddrPort{})
}

func (m mockConn) RemoteAddr() net.Addr {
	return net.TCPAddrFromAddrPort(netip.AddrPort{})
}

func (m mockConn) SetDeadline(t time.Time) error {
	return nil
}

func (m mockConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (m mockConn) SetWriteDeadline(t time.Time) error {
	return nil
}

//
//func TestServer_doCommand(t *testing.T) {
//	type args struct {
//		session *Session
//		rawLine string
//		conn    net.Conn
//		reader  *bufio.Reader
//	}
//	tests := []struct {
//		name string
//		args args
//	}{
//		{
//			name: "StatusTest",
//			args: args{
//				session: &Session{
//					Status: AUTHORIZED,
//
//				},
//				rawLine: `9.33 STATUS "Sent Messages" (MESSAGES UIDNEXT UIDVALIDITY UNSEEN)`,
//				conn:    &mockConn{},
//				reader:  &bufio.Reader{},
//			},
//		},
//		{
//			name: "StatusTest2",
//			args: args{
//				session: &Session{
//					Status: AUTHORIZED,
//				},
//				rawLine: `9.33 STATUS INBOX (MESSAGES UIDNEXT UIDVALIDITY UNSEEN)`,
//				conn:    &mockConn{},
//				reader:  &bufio.Reader{},
//			},
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			s := &Server{
//			}
//			s.doCommand(tt.args.session, tt.args.rawLine, tt.args.conn, tt.args.reader)
//		})
//	}
//}
