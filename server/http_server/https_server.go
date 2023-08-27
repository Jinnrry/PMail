package http_server

import (
	"bytes"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"io/fs"
	olog "log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"pmail/config"
	"pmail/controllers"
	"pmail/controllers/email"
	"pmail/dto"
	"pmail/dto/response"
	"pmail/i18n"
	"pmail/session"
	"time"
)

//go:embed dist/*
var local embed.FS

const HttpsPort = 443

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
	mux.HandleFunc("/attachments/", contextIterceptor(controllers.GetAttachments))
	mux.HandleFunc("/attachments/download/", contextIterceptor(controllers.Download))

	// go http server会打一堆没用的日志，写一个空的日志处理器，屏蔽掉日志输出
	nullLog := olog.New(&nullWrite{}, "", olog.Ldate)

	if config.Instance.HttpsEnabled != 2 {
		httpsServer = &http.Server{
			Addr:         fmt.Sprintf(":%d", HttpsPort),
			Handler:      session.Instance.LoadAndSave(mux),
			ReadTimeout:  time.Second * 60,
			WriteTimeout: time.Second * 60,
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

func genLogID() string {
	r := rand.New(rand.NewSource(time.Now().UnixMicro()))
	if ip == "" {
		ip = getLocalIP()
	}
	now := time.Now()
	timestamp := uint32(now.Unix())
	timeNano := now.UnixNano()
	pid := os.Getpid()
	b := bytes.Buffer{}

	b.WriteString(hex.EncodeToString(net.ParseIP(ip).To4()))
	b.WriteString(fmt.Sprintf("%x", timestamp&0xffffffff))
	b.WriteString(fmt.Sprintf("%04x", timeNano&0xffff))
	b.WriteString(fmt.Sprintf("%04x", pid&0xffff))
	b.WriteString(fmt.Sprintf("%06x", r.Int31n(1<<24)))
	b.WriteString("b0")

	return b.String()
}

// 注入context
func contextIterceptor(h controllers.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if w.Header().Get("Content-Type") == "" {
			w.Header().Set("Content-Type", "application/json")
		}

		ctx := &dto.Context{}
		ctx.Context = r.Context()
		ctx.SetValue(dto.LogID, genLogID())
		lang := r.Header.Get("Lang")
		if lang == "" {
			lang = "en"
		}
		ctx.Lang = lang

		if config.IsInit {
			user := cast.ToString(session.Instance.Get(ctx, "user"))
			if user != "" {
				_ = json.Unmarshal([]byte(user), &ctx.UserInfo)
			}
			if ctx.UserInfo == nil || ctx.UserInfo.ID == 0 {
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
