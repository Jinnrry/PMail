package http_server

import (
	"fmt"
	"io/fs"
	"net/http"
	"pmail/config"
	"pmail/controllers"
	"time"
)

var ip string

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
	if config.Instance != nil && config.Instance.HttpPort > 0 {
		HttpPort = config.Instance.HttpPort
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
	if err != nil {
		panic(err)
	}
}
