package framework

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"pmail/dto/parsemail"
	"pmail/utils/context"
)

type EmailHook interface {
	// SendBefore 邮件发送前的数据
	SendBefore(ctx *context.Context, email *parsemail.Email)
	// SendAfter 邮件发送后的数据，err是每个收信服务器的错误信息
	SendAfter(ctx *context.Context, email *parsemail.Email, err map[string]error)
	// ReceiveParseBefore 接收到邮件，解析之前的原始数据
	ReceiveParseBefore(ctx *context.Context, email *[]byte)
	// ReceiveParseAfter 接收到邮件，解析之后的结构化数据
	ReceiveParseAfter(ctx *context.Context, email *parsemail.Email)
}

type HookDTO struct {
	Ctx       *context.Context
	Email     *parsemail.Email
	EmailByte *[]byte
	ErrMap    map[string]error
}

type Plugin struct {
	hook EmailHook
}

func CreatePlugin(hook EmailHook) *Plugin {
	return &Plugin{
		hook: hook,
	}
}

func (p *Plugin) Run() {
	if len(os.Args) < 2 {
		panic("Command Params Error!")
	}
	mux := http.NewServeMux()

	mux.HandleFunc("/SendBefore", func(writer http.ResponseWriter, request *http.Request) {
		var hookDTO HookDTO
		body, _ := io.ReadAll(request.Body)
		err := json.Unmarshal(body, &hookDTO)
		if err != nil {
			log.Errorf("params error %+v", err)
			return
		}
		p.hook.SendBefore(hookDTO.Ctx, hookDTO.Email)
		body, _ = json.Marshal(hookDTO)
		writer.Write(body)
	})
	mux.HandleFunc("/SendAfter", func(writer http.ResponseWriter, request *http.Request) {

		var hookDTO HookDTO
		body, _ := io.ReadAll(request.Body)
		err := json.Unmarshal(body, &hookDTO)
		if err != nil {
			log.Errorf("params error %+v", err)
			return
		}
		p.hook.SendAfter(hookDTO.Ctx, hookDTO.Email, hookDTO.ErrMap)
		body, _ = json.Marshal(hookDTO)
		writer.Write(body)
	})
	mux.HandleFunc("/ReceiveParseBefore", func(writer http.ResponseWriter, request *http.Request) {

		var hookDTO HookDTO
		body, _ := io.ReadAll(request.Body)
		err := json.Unmarshal(body, &hookDTO)
		if err != nil {
			log.Errorf("params error %+v", err)
			return
		}
		p.hook.ReceiveParseBefore(hookDTO.Ctx, hookDTO.EmailByte)
		body, _ = json.Marshal(hookDTO)
		writer.Write(body)
	})
	mux.HandleFunc("/ReceiveParseAfter", func(writer http.ResponseWriter, request *http.Request) {

		var hookDTO HookDTO
		body, _ := io.ReadAll(request.Body)
		err := json.Unmarshal(body, &hookDTO)
		if err != nil {
			log.Errorf("params error %+v", err)
			return
		}
		p.hook.ReceiveParseAfter(hookDTO.Ctx, hookDTO.Email)
		body, _ = json.Marshal(hookDTO)
		writer.Write(body)
	})

	server := http.Server{
		Handler: mux,
	}

	unixListener, err := net.Listen("unix", getExePath()+"/"+os.Args[1])
	if err != nil {
		panic(err)
	}
	err = server.Serve(unixListener)
	if err != nil {
		panic(err)
	}
}

func getExePath() string {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exePath := filepath.Dir(ex)
	return exePath
}
