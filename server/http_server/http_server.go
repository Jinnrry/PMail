package http_server

import (
	"fmt"
	"io/fs"
	"net/http"
	"pmail/config"
	"pmail/controllers"
	"pmail/controllers/email"
	"pmail/session"
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
		httpServer = &http.Server{
			Addr:         fmt.Sprintf(":%d", HttpPort),
			Handler:      mux,
			ReadTimeout:  time.Second * 60,
			WriteTimeout: time.Second * 60,
		}
	} else {
		fe, err := fs.Sub(local, "dist")
		if err != nil {
			panic(err)
		}
		mux.Handle("/", http.FileServer(http.FS(fe)))
		mux.HandleFunc("/api/ping", contextIterceptor(controllers.Ping))
		mux.HandleFunc("/api/login", contextIterceptor(controllers.Login))
		mux.HandleFunc("/api/group", contextIterceptor(controllers.GetUserGroup))
		mux.HandleFunc("/api/group/list", contextIterceptor(controllers.GetUserGroupList))
		mux.HandleFunc("/api/group/add", contextIterceptor(controllers.AddGroup))
		mux.HandleFunc("/api/group/del", contextIterceptor(controllers.DelGroup))
		mux.HandleFunc("/api/email/list", contextIterceptor(email.EmailList))
		mux.HandleFunc("/api/email/del", contextIterceptor(email.EmailDelete))
		mux.HandleFunc("/api/email/read", contextIterceptor(email.MarkRead))
		mux.HandleFunc("/api/email/detail", contextIterceptor(email.EmailDetail))
		mux.HandleFunc("/api/email/move", contextIterceptor(email.Move))
		mux.HandleFunc("/api/email/send", contextIterceptor(email.Send))
		mux.HandleFunc("/api/settings/modify_password", contextIterceptor(controllers.ModifyPassword))
		mux.HandleFunc("/attachments/", contextIterceptor(controllers.GetAttachments))
		mux.HandleFunc("/attachments/download/", contextIterceptor(controllers.Download))
		httpServer = &http.Server{
			Addr:         fmt.Sprintf(":%d", HttpPort),
			Handler:      session.Instance.LoadAndSave(mux),
			ReadTimeout:  time.Second * 60,
			WriteTimeout: time.Second * 60,
		}
	}

	err := httpServer.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
