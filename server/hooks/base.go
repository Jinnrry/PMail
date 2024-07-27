package hooks

import (
	oContext "context"
	"encoding/json"
	"fmt"
	"github.com/Jinnrry/pmail/dto/parsemail"
	"github.com/Jinnrry/pmail/hooks/framework"
	"github.com/Jinnrry/pmail/models"
	"github.com/Jinnrry/pmail/utils/context"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// HookList
var HookList map[string]framework.EmailHook

type HookSender struct {
	httpc  http.Client
	name   string
	socket string
}

func (h *HookSender) ReceiveSaveAfter(ctx *context.Context, email *parsemail.Email, ue []*models.UserEmail) {
	log.WithContext(ctx).Debugf("[%s]Plugin ReceiveSaveAfter Start", h.name)

	dto := framework.HookDTO{
		Ctx:       ctx,
		Email:     email,
		UserEmail: ue,
	}
	body, _ := json.Marshal(dto)

	_, err := h.httpc.Post("http://plugin/ReceiveSaveAfter", "application/json", strings.NewReader(string(body)))
	if err != nil {
		log.WithContext(ctx).Errorf("[%s] Error! %v", h.name, err)
		return
	}

	log.WithContext(ctx).Debugf("[%s]Plugin ReceiveSaveAfter End", h.name)
}

func (h *HookSender) SendBefore(ctx *context.Context, email *parsemail.Email) {
	log.WithContext(ctx).Debugf("[%s]Plugin SendBefore Start", h.name)

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
	log.WithContext(ctx).Debugf("[%s]Plugin SendBefore End", h.name)

}

func (h *HookSender) SendAfter(ctx *context.Context, email *parsemail.Email, err map[string]error) {
	log.WithContext(ctx).Debugf("[%s]Plugin SendAfter Start", h.name)
	dto := framework.HookDTO{
		Ctx:    ctx,
		Email:  email,
		ErrMap: err,
	}
	body, _ := json.Marshal(dto)

	_, errL := h.httpc.Post("http://plugin/SendAfter", "application/json", strings.NewReader(string(body)))
	if errL != nil {
		log.WithContext(ctx).Errorf("[%s] Error! %v", h.name, errL)
		return
	}

	log.WithContext(ctx).Debugf("[%s]Plugin SendAfter End", h.name)

}

func (h *HookSender) ReceiveParseBefore(ctx *context.Context, email *[]byte) {
	log.WithContext(ctx).Debugf("[%s]Plugin ReceiveParseBefore Start", h.name)

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
	log.WithContext(ctx).Debugf("[%s]Plugin ReceiveParseBefore End", h.name)

}

func (h *HookSender) ReceiveParseAfter(ctx *context.Context, email *parsemail.Email) {
	log.WithContext(ctx).Debugf("[%s]Plugin ReceiveParseAfter Start", h.name)

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
	log.WithContext(ctx).Debugf("[%s]Plugin ReceiveParseAfter End", h.name)

}

// GetName 获取插件名称
func (h *HookSender) GetName(ctx *context.Context) string {

	dto := framework.HookDTO{
		Ctx: ctx,
	}
	body, _ := json.Marshal(dto)

	ret, errL := h.httpc.Post("http://plugin/GetName", "application/json", strings.NewReader(string(body)))
	if errL != nil {
		log.WithContext(ctx).Errorf("[%s] Error! %v", h.name, errL)
		return ""
	}

	body, _ = io.ReadAll(ret.Body)

	return string(body)
}

// SettingsHtml 插件页面
func (h *HookSender) SettingsHtml(ctx *context.Context, url string, requestData string) string {

	dto := framework.SettingsHtmlRequest{
		Ctx:         ctx,
		URL:         url,
		RequestData: requestData,
	}
	body, _ := json.Marshal(dto)

	ret, errL := h.httpc.Post("http://plugin/SettingsHtml", "application/json", strings.NewReader(string(body)))
	if errL != nil {
		log.WithContext(ctx).Errorf("[%s] Error! %v", h.name, errL)
		return ""
	}

	body, _ = io.ReadAll(ret.Body)

	return string(body)

}

func NewHookSender(socketPath string, name string, serverVersion string) *HookSender {
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

var processList []*os.Process

// Init 注册hook对象
func Init(serverVersion string) {

	HookList = map[string]framework.EmailHook{}
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

			//socketPath = "/PMail/server/hooks/spam_block/1555.socket"  //debug

			log.Infof("[%s] Plugin Load", info.Name())
			p, err := os.StartProcess(path, []string{
				info.Name(),
				fmt.Sprintf("%d.socket", pluginNo),
			}, procAttr)
			if err != nil {
				log.Errorf("Plug Load Error! %v", err)
				return nil
			}
			fmt.Printf("[%s] Plugin Start! PID:%d", info.Name(), p.Pid)
			processList = append(processList, p)

			pluginNo++

			go func() {
				stat, err := p.Wait()
				log.Errorf("[%s] Plugin Stop. Error:%v Stat:%v", info.Name(), err, stat.String())
				delete(HookList, info.Name())
				os.Remove(socketPath)
			}()

			loadSucc := false
			for i := 0; i < 5; i++ {
				time.Sleep(1 * time.Second)
				if _, err := os.Stat(socketPath); err == nil {
					loadSucc = true
					break
				}
				if i == 4 {
					log.Errorf(fmt.Sprintf("[%s] Start Fail!", info.Name()))
				}
			}
			if loadSucc {
				hk := NewHookSender(socketPath, info.Name(), serverVersion)
				hkName := hk.GetName(&context.Context{})
				HookList[hkName] = hk
				log.Infof("[%s] Plugin Load Success!", hkName)
			}

		}

		return nil
	})

}

func Stop() {
	log.Info("Plugin Stop")
	for _, process := range processList {
		process.Kill()
	}
}
