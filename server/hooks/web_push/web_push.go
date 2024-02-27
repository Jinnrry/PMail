package web_push

import (
	"bytes"
	"encoding/json"
	"net/http"
	"pmail/config"
	"pmail/dto/parsemail"
	"pmail/utils/context"

	log "github.com/sirupsen/logrus"
)

type WebPushHook struct {
	url   string
	token string
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

func (w *WebPushHook) ReceiveParseBefore(email []byte) {

}

func (w *WebPushHook) ReceiveParseAfter(email *parsemail.Email) {
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

	var ctx *context.Context = nil
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

func NewWebPushHook() *WebPushHook {

	ret := &WebPushHook{
		url:   config.Instance.WebPushUrl,
		token: config.Instance.WebPushToken,
	}
	return ret

}
