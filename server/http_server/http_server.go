package http_server

import (
	"fmt"
	"net/http"
	"pmail/controllers"
	"time"
)

const HttpPort = 80

// 这个服务是为了拦截http请求转发到https
var httpServer *http.Server

func HttpStop() {
	if httpServer != nil {
		httpServer.Close()
	}
}

func HttpStart() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", controllers.Interceptor)
	httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", HttpPort),
		Handler:      mux,
		ReadTimeout:  time.Second * 60,
		WriteTimeout: time.Second * 60,
	}

	err := httpServer.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
