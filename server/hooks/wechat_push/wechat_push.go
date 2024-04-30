package main

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"io"
	"net/http"
	"os"
	"pmail/config"
	"pmail/dto/parsemail"
	"pmail/hooks/framework"
	"pmail/utils/context"
	"strings"
	"time"
)

type accessTokenRes struct {
	AccessToken string `db:"access_token" json:"access_token"`
	ExpiresIn   int    `db:"expires_in" json:"expires_in"`
}

type WeChatPushHook struct {
	appId        string
	secret       string
	token        string
	tokenExpires int64
	templateId   string
	pushUser     string
	mainConfig   *config.Config
}

func (w *WeChatPushHook) ReceiveSaveAfter(ctx *context.Context, email *parsemail.Email) {
	if w.appId == "" || w.secret == "" || w.pushUser == "" {
		return
	}

	// 被标记为已读，或者是已删除，或是垃圾邮件 就不处理了
	if email.IsRead == 1 || email.Status == 3 || email.MessageId <= 0 {
		return
	}

	content := string(email.Text)

	if content == "" {
		content = email.Subject
	}

	w.sendUserMsg(nil, w.pushUser, content)
}

func (w *WeChatPushHook) SendBefore(ctx *context.Context, email *parsemail.Email) {

}

func (w *WeChatPushHook) SendAfter(ctx *context.Context, email *parsemail.Email, err map[string]error) {

}

func (w *WeChatPushHook) ReceiveParseBefore(ctx *context.Context, email *[]byte) {

}

func (w *WeChatPushHook) ReceiveParseAfter(ctx *context.Context, email *parsemail.Email) {}

func (w *WeChatPushHook) getWxAccessToken() string {
	if w.tokenExpires > time.Now().Unix() {
		return w.token
	}
	resp, err := http.Get(fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s", w.appId, w.secret))
	if err != nil {
		return ""
	}
	body, _ := io.ReadAll(resp.Body)
	var ret accessTokenRes
	_ = json.Unmarshal(body, &ret)
	if ret.AccessToken != "" {
		w.token = ret.AccessToken
		w.tokenExpires = time.Now().Unix() + cast.ToInt64(ret.ExpiresIn)
	}
	return ret.AccessToken
}

type sendMsgRequest struct {
	Touser      string   `db:"touser" json:"touser"`
	Template_id string   `db:"template_id" json:"template_id"`
	Url         string   `db:"url" json:"url"`
	Data        SendData `db:"data" json:"data"`
}
type SendData struct {
	Content DataItem `json:"Content"`
}
type DataItem struct {
	Value string `json:"value"`
	Color string `json:"color"`
}

func (w *WeChatPushHook) sendUserMsg(ctx *context.Context, userId string, content string) {

	url := w.mainConfig.WebDomain
	if w.mainConfig.HttpsEnabled > 1 {
		url = "http://" + url
	} else {
		url = "https://" + url
	}

	sendMsgReq, _ := json.Marshal(sendMsgRequest{
		Touser:      userId,
		Template_id: w.templateId,
		Url:         url,
		Data:        SendData{Content: DataItem{Value: content, Color: "#000000"}},
	})

	_, err := http.Post("https://api.weixin.qq.com/cgi-bin/message/template/send?access_token="+w.getWxAccessToken(), "application/json", strings.NewReader(string(sendMsgReq)))
	if err != nil {
		log.WithContext(ctx).Errorf("wechat push error %+v", err)
	}

}

type Config struct {
	WeChatPushAppId      string `json:"weChatPushAppId"`
	WeChatPushSecret     string `json:"weChatPushSecret"`
	WeChatPushTemplateId string `json:"weChatPushTemplateId"`
	WeChatPushUserId     string `json:"weChatPushUserId"`
}

func NewWechatPushHook() *WeChatPushHook {

	var cfgData []byte
	var err error

	cfgData, err = os.ReadFile("../../config/config.json")
	if err != nil {
		panic(err)
	}
	var mainConfig *config.Config
	err = json.Unmarshal(cfgData, &mainConfig)
	if err != nil {
		panic(err)
	}

	var pluginConfig *Config
	if _, err := os.Stat("./plugins/wechat_push_config.json"); err == nil {
		cfgData, err = os.ReadFile("./plugins/wechat_push_config.json")
		if err != nil {
			panic(err)
		}
		err = json.Unmarshal(cfgData, &pluginConfig)
		if err != nil {
			panic(err)
		}

	}

	appid := ""
	secret := ""
	templateId := ""
	userId := ""
	if pluginConfig != nil {
		appid = pluginConfig.WeChatPushAppId
		secret = pluginConfig.WeChatPushSecret
		templateId = pluginConfig.WeChatPushTemplateId
		userId = pluginConfig.WeChatPushUserId
	} else {
		appid = mainConfig.WeChatPushAppId
		secret = mainConfig.WeChatPushSecret
		templateId = mainConfig.WeChatPushTemplateId
		userId = mainConfig.WeChatPushUserId
	}

	ret := &WeChatPushHook{
		appId:      appid,
		secret:     secret,
		templateId: templateId,
		pushUser:   userId,
		mainConfig: mainConfig,
	}
	return ret

}

// 插件将以独立进程运行，因此需要主函数。
func main() {
	framework.CreatePlugin("wechat_push", NewWechatPushHook()).Run()
}
