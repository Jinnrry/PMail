package pop3_server

import (
	"github.com/Jinnrry/gopop"
	log "github.com/sirupsen/logrus"
	"pmail/config"
)

var instance *gopop.Server

func Start() {
	instance = gopop.NewPop3Server(110, config.Instance.Domain, false, action{})
	log.Infof("POP3 Server Start On Port :110")

	err := instance.Start()
	if err != nil {
		panic(err)
	}
}

func Stop() {
	instance.Stop()
}
