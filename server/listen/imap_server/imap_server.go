package imap_server

import (
	"crypto/rand"
	"crypto/tls"
	"github.com/Jinnrry/pmail/config"
	"github.com/Jinnrry/pmail/listen/imap_server/goimap"
	log "github.com/sirupsen/logrus"
	"time"
)

var instanceTLS *goimap.Server

// StarTLS 启动TLS端口监听，不加密的代码就懒得写了
func StarTLS() {
	crt, err := tls.LoadX509KeyPair(config.Instance.SSLPublicKeyPath, config.Instance.SSLPrivateKeyPath)
	if err != nil {
		panic(err)
	}
	tlsConfig := &tls.Config{}
	tlsConfig.Certificates = []tls.Certificate{crt}
	tlsConfig.Time = time.Now
	tlsConfig.Rand = rand.Reader
	instanceTLS = goimap.NewImapServer(993, "imap."+config.Instance.Domain, true, tlsConfig, action{})
	instanceTLS.ConnectAliveTime = 5 * time.Minute

	log.Infof("IMAP With TLS Server Start On Port :993")

	err = instanceTLS.Start()
	if err != nil {
		panic(err)
	}
}

func Stop() {
	if instanceTLS != nil {
		instanceTLS.Stop()
	}
}
