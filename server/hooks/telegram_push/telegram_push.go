package telegram_push

import (
	"encoding/json"
	"fmt"
	"net/http"
	"pmail/config"
	"pmail/dto"
	"pmail/dto/parsemail"
	"strings"

	log "github.com/sirupsen/logrus"
)

type TelegramPushHook struct {
	chatId       string
	botToken     string
	httpsEnabled int
	webDomain    string
}

func (w *TelegramPushHook) SendBefore(ctx *dto.Context, email *parsemail.Email) {

}

func (w *TelegramPushHook) SendAfter(ctx *dto.Context, email *parsemail.Email, err map[string]error) {

}

func (w *TelegramPushHook) ReceiveParseBefore(email []byte) {

}

func (w *TelegramPushHook) ReceiveParseAfter(email *parsemail.Email) {
	if w.chatId == "" || w.botToken == "" {
		return
	}

	w.sendUserMsg(nil, email)
}

type SendMessageRequest struct {
	ChatID      string      `json:"chat_id"`
	Text        string      `json:"text"`
	ReplyMarkup ReplyMarkup `json:"reply_markup"`
	ParseMode   string      `json:"parse_mode"`
}

type ReplyMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

type InlineKeyboardButton struct {
	Text string `json:"text"`
	URL  string `json:"url"`
}

func (w *TelegramPushHook) sendUserMsg(ctx *dto.Context, email *parsemail.Email) {
	url := w.webDomain
	if w.httpsEnabled > 1 {
		url = "http://" + url
	} else {
		url = "https://" + url
	}
	sendMsgReq, _ := json.Marshal(SendMessageRequest{
		ChatID:    w.chatId,
		Text:      fmt.Sprintf("📧<b>%s</b>&#60;%s&#62;\n\n%s", email.Subject, email.From.EmailAddress, string(email.Text)),
		ParseMode: "HTML",
		ReplyMarkup: ReplyMarkup{
			InlineKeyboard: [][]InlineKeyboardButton{
				{
					{
						Text: "查收邮件",
						URL:  url,
					},
				},
			},
		},
	})

	_, err := http.Post(fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", w.botToken), "application/json", strings.NewReader(string(sendMsgReq)))
	if err != nil {
		log.WithContext(ctx).Errorf("telegram push error %+v", err)
	}

}
func NewTelegramPushHook() *TelegramPushHook {
	ret := &TelegramPushHook{
		botToken:     config.Instance.TgBotToken,
		chatId:       config.Instance.TgChatId,
		webDomain:    config.Instance.WebDomain,
		httpsEnabled: config.Instance.HttpsEnabled,
	}
	return ret

}
