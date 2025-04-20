package main

import (
	"encoding/json"
	"fmt"
	"github.com/Jinnrry/pmail/config"
	"github.com/Jinnrry/pmail/dto/parsemail"
	"github.com/Jinnrry/pmail/hooks/framework"
	"github.com/Jinnrry/pmail/models"
	"github.com/Jinnrry/pmail/utils/context"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strings"
)

type TelegramPushHook struct {
	chatId       string
	botToken     string
	httpsEnabled int
	webDomain    string
}

func (w *TelegramPushHook) ReceiveSaveAfter(ctx *context.Context, email *parsemail.Email, ue []*models.UserEmail) {
	if w.chatId == "" || w.botToken == "" {
		return
	}

	for _, u := range ue {
		// 管理员（Uid=1）收到邮件且非已读、非已删除 触发通知
		if u.UserID == 1 && u.IsRead == 0 && u.Status == 0 && email.MessageId > 0 {
			w.sendUserMsg(nil, email)
		}
	}

}

// GetName 获取插件名称
func (w *TelegramPushHook) GetName(ctx *context.Context) string {
	return "TgPush"
}

// SettingsHtml 插件页面
func (w *TelegramPushHook) SettingsHtml(ctx *context.Context, url string, requestData string) string {
	return fmt.Sprintf(`
<div>
	 TG push No Settings Page
</div>
`)
}

func (w *TelegramPushHook) SendBefore(ctx *context.Context, email *parsemail.Email) {

}

func (w *TelegramPushHook) SendAfter(ctx *context.Context, email *parsemail.Email, err map[string]error) {

}

func (w *TelegramPushHook) ReceiveParseBefore(ctx *context.Context, email *[]byte) {

}

func (w *TelegramPushHook) ReceiveParseAfter(ctx *context.Context, email *parsemail.Email) {}

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

func (w *TelegramPushHook) sendUserMsg(ctx *context.Context, email *parsemail.Email) {
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

type Config struct {
	TgBotToken string `json:"tgBotToken"`
	TgChatId   string `json:"tgChatId"`
}

func NewTelegramPushHook() *TelegramPushHook {
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
	if _, err := os.Stat("./plugins/telegram_push_config.json"); err == nil {
		cfgData, err = os.ReadFile("./plugins/telegram_push_config.json")
		if err != nil {
			panic(err)
		}
		err = json.Unmarshal(cfgData, &pluginConfig)
		if err != nil {
			panic(err)
		}

	}

	token := ""
	chatID := ""
	if pluginConfig != nil {
		token = pluginConfig.TgBotToken
		chatID = pluginConfig.TgChatId
	} else {
		token = mainConfig.TgBotToken
		chatID = mainConfig.TgChatId
	}

	ret := &TelegramPushHook{
		botToken:     token,
		chatId:       chatID,
		webDomain:    mainConfig.WebDomain,
		httpsEnabled: mainConfig.HttpsEnabled,
	}
	return ret

}

func main() {
	framework.CreatePlugin("telegram_push", NewTelegramPushHook()).Run()
}
