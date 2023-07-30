package smtp_server

import (
	"pmail/dto/parsemail"
	"testing"
)

func TestSend(t *testing.T) {
	testInit()
	e := &parsemail.Email{
		From: &parsemail.User{
			Name:         "发送人",
			EmailAddress: "j@jinnrry.com",
		},
		To: []*parsemail.User{
			{"ok@jinnrry.com", "名"},
			{"ok@xjiangwei.cn", "字"},
		},
		Subject: "你好",
		Text:    []byte("这是Text"),
		HTML:    []byte("<div>这是Html</div>"),
	}
	Send(nil, e)
}
