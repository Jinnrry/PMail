package res_init

import (
	"encoding/json"
	"github.com/Jinnrry/pmail/config"
	"github.com/Jinnrry/pmail/db"
	"github.com/Jinnrry/pmail/dto/parsemail"
	"github.com/Jinnrry/pmail/hooks"
	"github.com/Jinnrry/pmail/http_server"
	"github.com/Jinnrry/pmail/pop3_server"
	"github.com/Jinnrry/pmail/services/setup/ssl"
	"github.com/Jinnrry/pmail/session"
	"github.com/Jinnrry/pmail/signal"
	"github.com/Jinnrry/pmail/smtp_server"
	"github.com/Jinnrry/pmail/utils/file"
	log "github.com/sirupsen/logrus"
	"os"
)

func Init(serverVersion string) {

	if !config.IsInit {
		dirInit()

		go http_server.SetupStart()
		<-signal.InitChan
		http_server.SetupStop()
	}

	for {
		config.Init()
		// 启动前检查一遍证书
		ssl.Update(false)
		parsemail.Init()
		err := db.Init(serverVersion)
		if err != nil {
			panic(err)
		}
		session.Init()
		hooks.Init(serverVersion)
		// smtp server start
		go smtp_server.Start()
		go smtp_server.StartWithTLS()
		// http server start
		go http_server.HttpsStart()
		go http_server.HttpStart()
		// pop3 server start
		go pop3_server.Start()
		go pop3_server.StartWithTls()

		configStr, _ := json.Marshal(config.Instance)
		log.Warnf("Config File Info:  %s", configStr)

		select {
		case <-signal.RestartChan:
			log.Infof("Server Restart!")
			smtp_server.Stop()
			http_server.HttpsStop()
			http_server.HttpStop()
			pop3_server.Stop()
			hooks.Stop()
		case <-signal.StopChan:
			log.Infof("Server Stop!")
			smtp_server.Stop()
			http_server.HttpsStop()
			http_server.HttpStop()
			pop3_server.Stop()
			hooks.Stop()
			return
		}

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
