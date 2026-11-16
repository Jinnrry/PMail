package smtp_server

import (
	"crypto/tls"
	"time"

	"github.com/Jinnrry/pmail/config"
	"github.com/emersion/go-smtp"
	log "github.com/sirupsen/logrus"
)

var instance *smtp.Server
var instanceTls *smtp.Server
var instanceTlsNew *smtp.Server

func newSMTPServer(addr string) *smtp.Server {
	server := smtp.NewServer(&Backend{})
	server.Addr = addr
	server.Domain = config.Instance.Domain
	server.ReadTimeout = 10 * time.Second
	server.WriteTimeout = 10 * time.Second
	server.MaxMessageBytes = 1024 * 1024 * 30
	server.MaxRecipients = 50
	// PLAIN and LOGIN credentials may only be sent over TLS, including on 587.
	server.AllowInsecureAuth = false
	return server
}

func StartWithTLSNew() {
	instanceTlsNew = newSMTPServer(":587")
	// Load the certificate and key
	cer, err := tls.LoadX509KeyPair(config.Instance.SSLPublicKeyPath, config.Instance.SSLPrivateKeyPath)
	if err != nil {
		log.Fatal(err)
		return
	}
	// Configure the TLS support for STARTTLS
	instanceTlsNew.TLSConfig = &tls.Config{Certificates: []tls.Certificate{cer}}

	log.Println("Starting Smtp With STARTTLS Server Port:", instanceTlsNew.Addr)
	// 587端口使用STARTTLS（先明文连接，再升级TLS），而非隐式TLS
	if err := instanceTlsNew.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func StartWithTLS() {
	instanceTls = newSMTPServer(":465")
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
	instance = newSMTPServer(":25")
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

	if instanceTlsNew != nil {
		instanceTlsNew.Close()
	}
}
