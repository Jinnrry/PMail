package main

import (
	"bytes"
	"encoding/json"
	"github.com/Jinnrry/pmail/config"
	"github.com/Jinnrry/pmail/dto/parsemail"
	"github.com/Jinnrry/pmail/hooks/framework"
	"github.com/Jinnrry/pmail/models"
	"github.com/Jinnrry/pmail/utils/context"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
)

type WebPushHook struct {
	url   string
	token string
}

func (w *WebPushHook) ReceiveSaveAfter(ctx *context.Context, email *parsemail.Email, ue []*models.UserEmail) {
	if w.url == "" {
		return
	}

	content := string(email.Text)

	if content == "" {
		content = email.Subject
	}

	webhookURL := w.url // 替换为您的 Webhook URL

	to := make([]string, len(email.To))
	for i, user := range email.To {
		to[i] = user.EmailAddress
	}

	data := EmailData{
		From:    email.From.EmailAddress,
		To:      to,
		Subject: email.Subject,
		Body:    content,
		Token:   w.token,
	}

	jsonData, err := json.Marshal(data)

	if err != nil {
		log.WithContext(ctx).Errorf("web push error %+v", err)
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.WithContext(ctx).Errorf("web push error %+v", err)
	}
	defer resp.Body.Close()
}

// EmailData 用于存储解析后的邮件数据
type EmailData struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Body    string   `json:"body"`
	Token   string   `json:"token"`
}

func (w *WebPushHook) SendBefore(ctx *context.Context, email *parsemail.Email) {

}

func (w *WebPushHook) SendAfter(ctx *context.Context, email *parsemail.Email, err map[string]error) {

}

func (w *WebPushHook) ReceiveParseBefore(ctx *context.Context, email *[]byte) {

}

func (w *WebPushHook) ReceiveParseAfter(ctx *context.Context, email *parsemail.Email) {
}

type Config struct {
	WebPushUrl   string `json:"webPushUrl"`
	WebPushToken string `json:"webPushToken"`
}

func NewWebPushHook() *WebPushHook {
	var cfgData []byte
	var err error

	cfgData, err = os.ReadFile("./config/config.json")
	if err != nil {
		panic(err)
	}
	var mainConfig *config.Config
	err = json.Unmarshal(cfgData, &mainConfig)
	if err != nil {
		panic(err)
	}

	var pluginConfig *Config
	if _, err := os.Stat("./plugins/web_push_config.json"); err == nil {
		cfgData, err = os.ReadFile("./plugins/web_push_config.json")
		if err != nil {
			panic(err)
		}
		err = json.Unmarshal(cfgData, &pluginConfig)
		if err != nil {
			panic(err)
		}

	}

	token := ""
	pushURL := ""
	if pluginConfig != nil {
		pushURL = pluginConfig.WebPushUrl
		token = pluginConfig.WebPushToken
	} else {
		pushURL = mainConfig.WebPushUrl
		token = mainConfig.WebPushToken
	}

	ret := &WebPushHook{
		url:   pushURL,
		token: token,
	}
	return ret

}

func main() {
	framework.CreatePlugin("web_push", NewWebPushHook()).Run()
}
