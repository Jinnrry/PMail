package smtp_server

import (
	"crypto/tls"
	"fmt"
	"time"

	"github.com/Jinnrry/pmail/config"
	"github.com/emersion/go-smtp"
	log "github.com/sirupsen/logrus"
)

var instance *smtp.Server
var instanceTls *smtp.Server

func StartWithTLS() {
	be := &Backend{}

	instanceTls = smtp.NewServer(be)

	if config.Instance.SMTPSPort == 0 {
		instanceTls.Addr = ":465"
	} else {
		instanceTls.Addr = fmt.Sprintf(":%d", config.Instance.SMTPSPort)
	}
	instanceTls.Domain = config.Instance.Domain
	instanceTls.ReadTimeout = 10 * time.Second
	instanceTls.WriteTimeout = 10 * time.Second
	instanceTls.MaxMessageBytes = 1024 * 1024 * 30
	instanceTls.MaxRecipients = 50
	instanceTls.AllowInsecureAuth = true
	// Load the certificate and key
	cer, err := tls.LoadX509KeyPair(config.Instance.SSLPublicKeyPath, config.Instance.SSLPrivateKeyPath)
	if err != nil {
		log.Fatal(err)
		return
	}
	// Configure the TLS support
	instanceTls.TLSConfig = &tls.Config{Certificates: []tls.Certificate{cer}}

	log.Println("Starting Smtp With SSL Server Port:", instanceTls.Addr)
	if err := instanceTls.ListenAndServeTLS(); err != nil {
		log.Fatal(err)
	}
}

func Start() {
	be := &Backend{}

	instance = smtp.NewServer(be)

	if config.Instance.SMTPPort == 0 {
		instance.Addr = ":25"
	} else {
		instance.Addr = fmt.Sprintf(":%d", config.Instance.SMTPPort)
	}
	instance.Domain = config.Instance.Domain
	instance.ReadTimeout = 10 * time.Second
	instance.WriteTimeout = 10 * time.Second
	instance.MaxMessageBytes = 1024 * 1024 * 30
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

	log.Println("Starting Smtp Server Port:", instance.Addr)
	if err := instance.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func Stop() {
	if instance != nil {
		instance.Close()
	}
	if instanceTls != nil {
		instanceTls.Close()
	}
}
