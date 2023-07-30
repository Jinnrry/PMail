package hooks

import (
	"pmail/dto"
	"pmail/dto/parsemail"
	"pmail/hooks/wechat_push"
)

type EmailHook interface {
	// SendBefore 邮件发送前的数据
	SendBefore(ctx *dto.Context, email *parsemail.Email)
	// SendAfter 邮件发送后的数据，err是每个收信服务器的错误信息
	SendAfter(ctx *dto.Context, email *parsemail.Email, err map[string]error)
	// ReceiveParseBefore 接收到邮件，解析之前的原始数据
	ReceiveParseBefore(email []byte)
	// ReceiveParseAfter 接收到邮件，解析之后的结构化数据
	ReceiveParseAfter(email *parsemail.Email)
}

// HookList
var HookList []EmailHook

// Init 这里注册hook对象
func Init() {
	HookList = []EmailHook{
		wechat_push.NewWechatPushHook(),
	}
}
