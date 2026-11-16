package smtp_server

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"net"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"strings"
	"testing"
	"time"

	"github.com/Jinnrry/pmail/models"
	"github.com/Jinnrry/pmail/utils/context"
	"github.com/Jinnrry/pmail/utils/password"
	log "github.com/sirupsen/logrus"
)

func TestSMTPAuthenticationRequiresTLS(t *testing.T) {
	engine, _ := newAuditTestEngine(t)
	if _, err := engine.Insert(&models.User{Account: "tls-user", Password: password.Encode("test-password")}); err != nil {
		t.Fatal(err)
	}
	// Reuse Go's self-signed test certificate without production certificate files.
	https := httptest.NewTLSServer(http.NotFoundHandler())
	certificates := https.TLS.Certificates
	https.Close()
	server := newSMTPServer("127.0.0.1:0")
	server.TLSConfig = &tls.Config{Certificates: certificates}
	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- server.Serve(listener) }()
	t.Cleanup(func() { server.Close(); <-done })

	for _, mechanism := range []string{"PLAIN", "LOGIN"} {
		t.Run(mechanism, func(t *testing.T) {
			conn, err := net.Dial("tcp", listener.Addr().String())
			if err != nil {
				t.Fatal(err)
			}
			defer conn.Close()
			conn.SetDeadline(time.Now().Add(5 * time.Second))
			client := textproto.NewConn(conn)
			if _, _, err := client.ReadResponse(220); err != nil {
				t.Fatal(err)
			}
			command := func(line string, code int) string {
				t.Helper()
				if err := client.PrintfLine("%s", line); err != nil {
					t.Fatal(err)
				}
				_, message, err := client.ReadResponse(code)
				if err != nil {
					t.Fatal(err)
				}
				return message
			}
			if caps := command("EHLO localhost", 250); strings.Contains(caps, "AUTH") {
				t.Fatal("advertised AUTH before TLS")
			}
			command("AUTH "+mechanism, 523)
			command("STARTTLS", 220)
			secure := tls.Client(conn, &tls.Config{InsecureSkipVerify: true}) // test certificate only
			if err := secure.Handshake(); err != nil {
				t.Fatal(err)
			}
			client = textproto.NewConn(secure)
			if caps := command("EHLO localhost", 250); !strings.Contains(caps, "AUTH") {
				t.Fatal("missing AUTH after TLS")
			}
			encode := func(s string) string { return base64.StdEncoding.EncodeToString([]byte(s)) }
			if mechanism == "PLAIN" {
				command("AUTH PLAIN "+encode("\x00tls-user\x00test-password"), 235)
			} else {
				command("AUTH LOGIN", 334)
				command(encode("tls-user"), 334)
				command(encode("test-password"), 235)
			}
			command("QUIT", 221)
		})
	}
}

func TestSMTPAuthenticationDoesNotLogCredentials(t *testing.T) {
	engine, _ := newAuditTestEngine(t)
	const good = "successful-login-secret"
	const bad = "failed-login-secret"
	encoded := password.Encode(good)
	if _, err := engine.Insert(&models.User{Account: "log-user", Password: encoded}); err != nil {
		t.Fatal(err)
	}
	logger := log.StandardLogger()
	oldOutput, oldLevel := logger.Out, logger.GetLevel()
	var output bytes.Buffer
	logger.SetOutput(&output)
	logger.SetLevel(log.DebugLevel)
	defer func() { logger.SetOutput(oldOutput); logger.SetLevel(oldLevel) }()
	for _, pwd := range []string{good, bad} {
		s := &Session{Ctx: &context.Context{}}
		err := s.AuthPlain("log-user", pwd)
		if (err == nil) != (pwd == good) {
			t.Fatalf("unexpected authentication result: %v", err)
		}
	}
	for _, secret := range []string{good, bad, encoded} {
		if strings.Contains(output.String(), secret) {
			t.Error("authentication log contains credentials")
		}
	}
}
