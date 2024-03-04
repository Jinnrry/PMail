package hooks

import (
	oContext "context"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"pmail/dto/parsemail"
	"pmail/hooks/framework"
	"pmail/utils/context"
	"strings"
	"time"
)

// HookList
var HookList []framework.EmailHook

type HookSender struct {
	httpc  http.Client
	name   string
	socket string
}

func (h *HookSender) SendBefore(ctx *context.Context, email *parsemail.Email) {

	dto := framework.HookDTO{
		Ctx:   ctx,
		Email: email,
	}
	body, _ := json.Marshal(dto)

	ret, err := h.httpc.Post("http://plugin/SendBefore", "application/json", strings.NewReader(string(body)))
	if err != nil {
		log.WithContext(ctx).Errorf("[%s] Error! %v", h.name, err)
		return
	}

	body, _ = io.ReadAll(ret.Body)
	json.Unmarshal(body, &dto)

	ctx = dto.Ctx
	email = dto.Email
}

func (h *HookSender) SendAfter(ctx *context.Context, email *parsemail.Email, err map[string]error) {
	dto := framework.HookDTO{
		Ctx:    ctx,
		Email:  email,
		ErrMap: err,
	}
	body, _ := json.Marshal(dto)

	ret, errL := h.httpc.Post("http://plugin/SendAfter", "application/json", strings.NewReader(string(body)))
	if errL != nil {
		log.WithContext(ctx).Errorf("[%s] Error! %v", h.name, errL)
		return
	}

	body, _ = io.ReadAll(ret.Body)
	json.Unmarshal(body, &dto)

	ctx = dto.Ctx
	email = dto.Email
	err = dto.ErrMap
}

func (h *HookSender) ReceiveParseBefore(ctx *context.Context, email *[]byte) {
	dto := framework.HookDTO{
		Ctx:       ctx,
		EmailByte: email,
	}
	body, _ := json.Marshal(dto)

	ret, errL := h.httpc.Post("http://plugin/ReceiveParseBefore", "application/json", strings.NewReader(string(body)))
	if errL != nil {
		log.WithContext(ctx).Errorf("[%s] Error! %v", h.name, errL)
		return
	}

	body, _ = io.ReadAll(ret.Body)
	json.Unmarshal(body, &dto)

	ctx = dto.Ctx
	email = dto.EmailByte
}

func (h *HookSender) ReceiveParseAfter(ctx *context.Context, email *parsemail.Email) {
	dto := framework.HookDTO{
		Ctx:   ctx,
		Email: email,
	}
	body, _ := json.Marshal(dto)

	ret, errL := h.httpc.Post("http://plugin/ReceiveParseAfter", "application/json", strings.NewReader(string(body)))
	if errL != nil {
		log.WithContext(ctx).Errorf("[%s] Error! %v", h.name, errL)
		return
	}

	body, _ = io.ReadAll(ret.Body)
	json.Unmarshal(body, &dto)

	ctx = dto.Ctx
	email = dto.Email
}

func NewHookSender(socketPath string, name string) *HookSender {
	httpc := http.Client{
		Timeout: time.Second * 10,
		Transport: &http.Transport{
			DialContext: func(ctx oContext.Context, network, addr string) (net.Conn, error) {
				return net.Dial("unix", socketPath)
			},
		},
	}
	return &HookSender{
		httpc:  httpc,
		socket: socketPath,
		name:   name,
	}
}

// Init 注册hook对象
func Init() {

	env := os.Environ()
	procAttr := &os.ProcAttr{
		Env: env,
		Files: []*os.File{
			os.Stdin,
			os.Stdout,
			os.Stderr,
		},
	}

	root := "./plugins"

	pluginNo := 1
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info != nil && !info.IsDir() && (!strings.Contains(info.Name(), ".") || strings.Contains(info.Name(), ".exe")) {

			socketPath := fmt.Sprintf("%s/%d.socket", root, pluginNo)

			os.Remove(socketPath)

			log.Infof("[%s] Plugin Load", info.Name())
			p, err := os.StartProcess(path, []string{
				info.Name(),
				fmt.Sprintf("%d.socket", pluginNo),
			}, procAttr)
			if err != nil {
				panic(err)
			}
			fmt.Printf("[%s] Plugin Start! PID:%d", info.Name(), p.Pid)

			pluginNo++

			HookList = append(HookList, NewHookSender(socketPath, info.Name()))

			go func() {
				stat, err := p.Wait()
				log.Errorf("[%s] Plugin Stop. Error:%v Stat:%v", info.Name(), err, stat.String())
			}()

			for i := 0; i < 5; i++ {
				time.Sleep(1 * time.Second)
				if _, err := os.Stat(socketPath); err == nil {
					break
				}
				if i == 4 {
					panic(fmt.Sprintf("[%s] Start Fail!", info.Name()))
				}
			}

		}

		return nil
	})

}
