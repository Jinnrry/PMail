package wechat_push

import (
	"pmail/config"
	"pmail/dto/parsemail"
	"testing"
)

func testInit() {

	config.Init()

}
func TestWeChatPushHook_ReceiveParseAfter(t *testing.T) {
	testInit()

	w := NewWechatPushHook()
	w.ReceiveParseAfter(&parsemail.Email{Subject: "标题", Text: []byte("文本内容")})
}
