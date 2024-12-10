package http_server

import (
	"flag"
	"fmt"
	"github.com/Jinnrry/pmail/config"
	"github.com/Jinnrry/pmail/controllers"
	"github.com/Jinnrry/pmail/utils/ip"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"io/fs"
	"net/http"
	"os"
	"strings"
	"time"
)

// 项目初始化引导用的服务，初始化引导结束后即退出
var setupServer *http.Server

func SetupStart() {
	mux := http.NewServeMux()
	fe, err := fs.Sub(local, "dist")
	if err != nil {
		panic(err)
	}
	mux.Handle("/", http.FileServer(http.FS(fe)))
	mux.HandleFunc("/api/", contextIterceptor(controllers.Setup))
	// 挑战请求类似这样 /.well-known/acme-challenge/QPyMAyaWw9s5JvV1oruyqWHG7OqkHMJEHPoUz2046KM
	mux.HandleFunc("/.well-known/", controllers.AcmeChallenge)

	HttpPort := 80
	flag.IntVar(&HttpPort, "p", 80, "初始化阶段Http服务端口")
	flag.Parse()

	if HttpPort == 80 {
		envs := os.Environ()
		for _, env := range envs {
			if strings.HasPrefix(env, "setup_port=") {
				HttpPort = cast.ToInt(strings.TrimSpace(strings.ReplaceAll(env, "setup_port=", "")))
			}
		}
	}

	if HttpPort <= 0 || HttpPort > 65535 {
		HttpPort = 80
	}

	config.Instance.SetSetupPort(HttpPort)
	log.Infof("HttpServer Start On Port :%d", HttpPort)
	if HttpPort == 80 {
		log.Infof("Please click http://%s to continue.\n", ip.GetIp())
	} else {
		log.Infof("Please click http://%s:%d to continue.", ip.GetIp(), HttpPort)
	}

	setupServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", HttpPort),
		Handler:      mux,
		ReadTimeout:  time.Second * 60,
		WriteTimeout: time.Second * 60,
	}
	err = setupServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}

func SetupStop() {
	err := setupServer.Close()
	log.Infof("Setup End!")
	if err != nil {
		panic(err)
	}
}
