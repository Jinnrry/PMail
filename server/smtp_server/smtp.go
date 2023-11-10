package smtp_server

import (
	"crypto/tls"
	"database/sql"
	"github.com/emersion/go-smtp"
	log "github.com/sirupsen/logrus"
	"net"
	"pmail/config"
	"pmail/db"
	"pmail/models"
	"pmail/utils/context"
	"pmail/utils/errors"
	"pmail/utils/id"
	"pmail/utils/password"
	"strings"
	"time"
)

// The Backend implements SMTP server methods.
type Backend struct{}

func (bkd *Backend) NewSession(conn *smtp.Conn) (smtp.Session, error) {

	remoteAddress := conn.Conn().RemoteAddr()
	ctx := &context.Context{}
	ctx.SetValue(context.LogID, id.GenLogID())
	log.WithContext(ctx).Debugf("新SMTP连接")

	return &Session{
		RemoteAddress: remoteAddress,
		Ctx:           ctx,
	}, nil
}

// A Session is returned after EHLO.
type Session struct {
	RemoteAddress net.Addr
	User          string
	From          string
	To            []string
	Ctx           *context.Context
}

func (s *Session) AuthPlain(username, pwd string) error {
	s.User = username

	var user models.User

	encodePwd := password.Encode(pwd)

	infos := strings.Split(username, "@")
	if len(infos) > 1 {
		username = infos[0]
	}

	err := db.Instance.Get(&user, db.WithContext(s.Ctx, "select * from user where account =? and password =?"),
		username, encodePwd)
	if err != nil && err != sql.ErrNoRows {
		log.Errorf("%+v", err)
	}

	if user.ID > 0 {
		s.Ctx.UserAccount = user.Account
		s.Ctx.UserID = user.ID
		s.Ctx.UserName = user.Name
		return nil
	}

	log.WithContext(s.Ctx).Debugf("登陆错误%s %s", username, pwd)
	return errors.New("password error")
}

func (s *Session) Mail(from string, opts *smtp.MailOptions) error {
	s.From = from
	return nil
}

func (s *Session) Rcpt(to string) error {
	s.To = append(s.To, to)
	return nil
}

func (s *Session) Reset() {}

func (s *Session) Logout() error {
	return nil
}

var instance *smtp.Server

func Start() {
	be := &Backend{}

	instance = smtp.NewServer(be)

	instance.Addr = ":25"
	instance.Domain = config.Instance.Domain
	instance.ReadTimeout = 10 * time.Second
	instance.AuthDisabled = false
	instance.WriteTimeout = 10 * time.Second
	instance.MaxMessageBytes = 1024 * 1024
	instance.MaxRecipients = 50
	// force TLS for auth
	instance.AllowInsecureAuth = false
	// Load the certificate and key
	cer, err := tls.LoadX509KeyPair(config.Instance.SSLPublicKeyPath, config.Instance.SSLPrivateKeyPath)
	if err != nil {
		log.Fatal(err)
		return
	}
	// Configure the TLS support
	instance.TLSConfig = &tls.Config{Certificates: []tls.Certificate{cer}}

	log.Println("Starting server at", instance.Addr)
	if err := instance.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func Stop() {
	if instance != nil {
		instance.Close()
	}
}
