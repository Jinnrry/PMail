package wechat_push

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"io"
	"net/http"
	"pmail/config"
	"pmail/dto/parsemail"
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
}

func (w *WeChatPushHook) SendBefore(ctx *context.Context, email *parsemail.Email) {

}

func (w *WeChatPushHook) SendAfter(ctx *context.Context, email *parsemail.Email, err map[string]error) {

}

func (w *WeChatPushHook) ReceiveParseBefore(email []byte) {

}

func (w *WeChatPushHook) ReceiveParseAfter(email *parsemail.Email) {
	if w.appId == "" || w.secret == "" || w.pushUser == "" {
		return
	}

	content := string(email.Text)

	if content == "" {
		content = email.Subject
	}

	w.sendUserMsg(nil, w.pushUser, content)
}

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

	url := config.Instance.WebDomain
	if config.Instance.HttpsEnabled > 1 {
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
func NewWechatPushHook() *WeChatPushHook {

	ret := &WeChatPushHook{
		appId:      config.Instance.WeChatPushAppId,
		secret:     config.Instance.WeChatPushSecret,
		templateId: config.Instance.WeChatPushTemplateId,
		pushUser:   config.Instance.WeChatPushUserId,
	}
	return ret

}
