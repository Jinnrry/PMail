package res_init

import (
	"os"
	"pmail/config"
	"pmail/db"
	"pmail/dto/parsemail"
	"pmail/hooks"
	"pmail/http_server"
	"pmail/session"
	"pmail/smtp_server"
	"pmail/utils/file"
)

func Init() {
	config.Init()

	if config.IsInit {
		parsemail.Init()
		db.Init()
		session.Init()
		hooks.Init()
		// smtp server start
		go smtp_server.Start()
		// http server start
		go http_server.Start()
	} else {
		dirInit()
		go http_server.SetupStart()
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
