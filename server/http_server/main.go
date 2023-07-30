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
	"math/rand"
	"net"
	"net/http"
	"os"
	"pmail/controllers"
	"pmail/controllers/email"
	"pmail/dto"
	"pmail/dto/response"
	"pmail/session"
	"time"
)

//go:embed dist/*
var local embed.FS

var ip string

const HttpPort = 80

func Start() {
	log.Infof("Http Server Start at :%d", HttpPort)

	mux := http.NewServeMux()

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

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", HttpPort),
		Handler:      session.Instance.LoadAndSave(mux),
		ReadTimeout:  time.Second * 60,
		WriteTimeout: time.Second * 60,
	}

	//err := server.ListenAndServeTLS( "config/ssl/public.crt", "config/ssl/private.key", nil)
	err = server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

func getLocalIP() string {
	ip := "127.0.0.1"
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ip
	}
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ip = ipnet.IP.String()
				break
			}
		}
	}

	return ip
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

		user := cast.ToString(session.Instance.Get(ctx, "user"))
		if user != "" {
			_ = json.Unmarshal([]byte(user), &ctx.UserInfo)
		}
		if ctx.UserInfo == nil || ctx.UserInfo.ID == 0 {
			if r.URL.Path != "/api/ping" && r.URL.Path != "/api/login" {
				response.NewErrorResponse(response.NeedLogin, "登陆已失效！", "").FPrint(w)
				return
			}
		}
		h(ctx, w, r)
	}
}
