package http_server

import (
	"fmt"
	"io/fs"
	"net/http"
	"pmail/config"
	"pmail/controllers"
	"pmail/controllers/email"
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

	if config.Instance.HttpsEnabled != 2 {
		mux.HandleFunc("/", controllers.Interceptor)
	} else {
		fe, err := fs.Sub(local, "dist")
		if err != nil {
			panic(err)
		}
		mux.Handle("/", http.FileServer(http.FS(fe)))
		mux.HandleFunc("/api/ping", contextIterceptor(controllers.Ping))
		mux.HandleFunc("/api/login", contextIterceptor(controllers.Login))
		mux.HandleFunc("/api/group", contextIterceptor(controllers.GetUserGroup))
		mux.HandleFunc("/api/email/list", contextIterceptor(email.EmailList))
		mux.HandleFunc("/api/email/detail", contextIterceptor(email.EmailDetail))
		mux.HandleFunc("/api/email/send", contextIterceptor(email.Send))
		mux.HandleFunc("/api/settings/modify_password", contextIterceptor(controllers.ModifyPassword))
		mux.HandleFunc("/attachments/", contextIterceptor(controllers.GetAttachments))
		mux.HandleFunc("/attachments/download/", contextIterceptor(controllers.Download))
	}

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
