package smtp_server

import (
	"crypto/tls"
	"github.com/emersion/go-smtp"
	log "github.com/sirupsen/logrus"
	"net"
	"pmail/config"
	"time"
)

// The Backend implements SMTP server methods.
type Backend struct{}

func (bkd *Backend) NewSession(conn *smtp.Conn) (smtp.Session, error) {
	remoteAddress := conn.Conn().RemoteAddr()

	return &Session{
		RemoteAddress: remoteAddress,
	}, nil
}

// A Session is returned after EHLO.
type Session struct {
	RemoteAddress net.Addr
}

func (s *Session) AuthPlain(username, password string) error {
	return nil
}

func (s *Session) Mail(from string, opts *smtp.MailOptions) error {
	return nil
}

func (s *Session) Rcpt(to string) error {
	return nil
}

func (s *Session) Reset() {}

func (s *Session) Logout() error {
	return nil
}

func Start() {
	be := &Backend{}

	s := smtp.NewServer(be)

	s.Addr = ":25"
	s.Domain = config.Instance.Domain
	s.ReadTimeout = 10 * time.Second
	s.WriteTimeout = 10 * time.Second
	s.MaxMessageBytes = 1024 * 1024
	s.MaxRecipients = 50
	// force TLS for auth
	s.AllowInsecureAuth = false
	// Load the certificate and key
	cer, err := tls.LoadX509KeyPair(config.Instance.SSLPublicKeyPath, config.Instance.SSLPrivateKeyPath)
	if err != nil {
		log.Fatal(err)
		return
	}
	// Configure the TLS support
	s.TLSConfig = &tls.Config{Certificates: []tls.Certificate{cer}}

	log.Println("Starting server at", s.Addr)
	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
