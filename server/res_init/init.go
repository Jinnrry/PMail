package res_init

import (
	log "github.com/sirupsen/logrus"
	"os"
	"pmail/config"
	"pmail/db"
	"pmail/dto/parsemail"
	"pmail/hooks"
	"pmail/http_server"
	"pmail/session"
	"pmail/signal"
	"pmail/smtp_server"
	"pmail/utils/file"
)

func Init() {

	if !config.IsInit {
		dirInit()

		log.Infof("Please click http://127.0.0.1 to continue.\n")
		go http_server.SetupStart()
		<-signal.InitChan
		http_server.SetupStop()
	}

	for {
		config.Init()
		parsemail.Init()
		err := db.Init()
		if err != nil {
			panic(err)
		}
		session.Init()
		hooks.Init()
		// smtp server start
		go smtp_server.Start()
		// http server start
		go http_server.HttpsStart()
		go http_server.HttpStart()
		<-signal.RestartChan
		log.Infof("Server Restart!")
		smtp_server.Stop()
		http_server.HttpsStop()
		http_server.HttpStop()
	}

}

func dirInit() {
	if !file.PathExist("./config") {
		err := os.MkdirAll("./config", 0744)
		if err != nil {
			panic(err)
		}
	}

	if !file.PathExist("./config/dkim") {
		err := os.MkdirAll("./config/dkim", 0744)
		if err != nil {
			panic(err)
		}
	}

	if !file.PathExist("./config/ssl") {
		err := os.MkdirAll("./config/ssl", 0744)
		if err != nil {
			panic(err)
		}
	}
}
