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
	"time"
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

// HookDTO PMail 主程序和插件通信的结构体
type HookDTO struct {
	ServerVersion string           // 服务端程序版本
	Ctx           *context.Context // 上下文
	Email         *parsemail.Email // 邮件内容
	EmailByte     *[]byte          // 未解析前的文件内容
	ErrMap        map[string]error // 错误信息
}

type Plugin struct {
	name string
	hook EmailHook
}

func CreatePlugin(name string, hook EmailHook) *Plugin {
	return &Plugin{
		name: name,
		hook: hook,
	}
}

func (p *Plugin) Run() {
	if len(os.Args) < 2 {
		panic("Command Params Error!")
	}
	mux := http.NewServeMux()

	mux.HandleFunc("/SendBefore", func(writer http.ResponseWriter, request *http.Request) {
		log.Debugf("[%s] SendBefore Start", p.name)
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
		log.Debugf("[%s] SendBefore End", p.name)
	})
	mux.HandleFunc("/SendAfter", func(writer http.ResponseWriter, request *http.Request) {
		log.Debugf("[%s] SendAfter Start", p.name)

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
		log.Debugf("[%s] SendAfter End", p.name)

	})
	mux.HandleFunc("/ReceiveParseBefore", func(writer http.ResponseWriter, request *http.Request) {
		log.Debugf("[%s] ReceiveParseBefore Start", p.name)
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
		log.Debugf("[%s] ReceiveParseBefore End", p.name)
	})
	mux.HandleFunc("/ReceiveParseAfter", func(writer http.ResponseWriter, request *http.Request) {
		log.Debugf("[%s] ReceiveParseAfter Start", p.name)
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
		log.Debugf("[%s] ReceiveParseAfter End", p.name)
	})

	server := http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		Handler:      mux,
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
