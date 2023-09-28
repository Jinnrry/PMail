package http_server

import (
	"embed"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"io/fs"
	olog "log"
	"net/http"
	"pmail/config"
	"pmail/controllers"
	"pmail/controllers/email"
	"pmail/dto/response"
	"pmail/i18n"
	"pmail/models"
	"pmail/session"
	"pmail/utils/context"
	"pmail/utils/id"
	"time"
)

//go:embed dist/*
var local embed.FS

var httpsServer *http.Server

type nullWrite struct {
}

func (w *nullWrite) Write(p []byte) (int, error) {
	return len(p), nil
}

func HttpsStart() {
	log.Infof("Http Server Start")

	mux := http.NewServeMux()

	fe, err := fs.Sub(local, "dist")
	if err != nil {
		panic(err)
	}
	mux.Handle("/", http.FileServer(http.FS(fe)))
	// 挑战请求类似这样 /.well-known/acme-challenge/QPyMAyaWw9s5JvV1oruyqWHG7OqkHMJEHPoUz2046KM
	mux.HandleFunc("/.well-known/", controllers.AcmeChallenge)
	mux.HandleFunc("/api/ping", contextIterceptor(controllers.Ping))
	mux.HandleFunc("/api/login", contextIterceptor(controllers.Login))
	mux.HandleFunc("/api/group", contextIterceptor(controllers.GetUserGroup))
	mux.HandleFunc("/api/group/list", contextIterceptor(controllers.GetUserGroupList))
	mux.HandleFunc("/api/group/add", contextIterceptor(controllers.AddGroup))
	mux.HandleFunc("/api/group/del", contextIterceptor(controllers.DelGroup))
	mux.HandleFunc("/api/email/list", contextIterceptor(email.EmailList))
	mux.HandleFunc("/api/email/read", contextIterceptor(email.MarkRead))
	mux.HandleFunc("/api/email/del", contextIterceptor(email.EmailDelete))
	mux.HandleFunc("/api/email/detail", contextIterceptor(email.EmailDetail))
	mux.HandleFunc("/api/email/send", contextIterceptor(email.Send))
	mux.HandleFunc("/api/email/move", contextIterceptor(email.Move))
	mux.HandleFunc("/api/settings/modify_password", contextIterceptor(controllers.ModifyPassword))
	mux.HandleFunc("/api/rule/get", contextIterceptor(controllers.GetRule))
	mux.HandleFunc("/api/rule/add", contextIterceptor(controllers.UpsertRule))
	mux.HandleFunc("/api/rule/update", contextIterceptor(controllers.UpsertRule))
	mux.HandleFunc("/api/rule/del", contextIterceptor(controllers.DelRule))
	mux.HandleFunc("/attachments/", contextIterceptor(controllers.GetAttachments))
	mux.HandleFunc("/attachments/download/", contextIterceptor(controllers.Download))

	// go http server会打一堆没用的日志，写一个空的日志处理器，屏蔽掉日志输出
	nullLog := olog.New(&nullWrite{}, "", olog.Ldate)

	HttpsPort := 443
	if config.Instance.HttpsPort > 0 {
		HttpsPort = config.Instance.HttpsPort
	}

	if config.Instance.HttpsEnabled != 2 {
		httpsServer = &http.Server{
			Addr:         fmt.Sprintf(":%d", HttpsPort),
			Handler:      session.Instance.LoadAndSave(mux),
			ReadTimeout:  time.Second * 90,
			WriteTimeout: time.Second * 90,
			ErrorLog:     nullLog,
		}
		err = httpsServer.ListenAndServeTLS("config/ssl/public.crt", "config/ssl/private.key")
		if err != nil {
			panic(err)
		}
	}
}

func HttpsStop() {
	if httpsServer != nil {
		httpsServer.Close()
	}
}

// 注入context
func contextIterceptor(h controllers.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if w.Header().Get("Content-Type") == "" {
			w.Header().Set("Content-Type", "application/json")
		}

		ctx := &context.Context{}
		ctx.Context = r.Context()
		ctx.SetValue(context.LogID, id.GenLogID())
		lang := r.Header.Get("Lang")
		if lang == "" {
			lang = "en"
		}
		ctx.Lang = lang

		if config.IsInit {
			user := cast.ToString(session.Instance.Get(ctx, "user"))
			var userInfo *models.User
			if user != "" {
				_ = json.Unmarshal([]byte(user), &userInfo)
			}
			if userInfo != nil && userInfo.ID > 0 {
				ctx.UserID = userInfo.ID
				ctx.UserName = userInfo.Name
				ctx.UserAccount = userInfo.Account
			}

			if ctx.UserID == 0 {
				if r.URL.Path != "/api/ping" && r.URL.Path != "/api/login" {
					response.NewErrorResponse(response.NeedLogin, i18n.GetText(ctx.Lang, "login_exp"), "").FPrint(w)
					return
				}
			}
		} else if r.URL.Path != "/api/setup" {
			response.NewErrorResponse(response.NeedSetup, "", "").FPrint(w)
			return
		}
		h(ctx, w, r)
	}
}
