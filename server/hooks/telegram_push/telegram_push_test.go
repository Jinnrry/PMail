package telegram_push

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

	w := NewTelegramPushHook()
	w.ReceiveParseAfter(&parsemail.Email{Subject: "标题", Text: []byte("文本内容"), From: &parsemail.User{
		EmailAddress: "hello@gmail.com",
	}})
}
