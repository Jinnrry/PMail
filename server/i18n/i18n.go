package i18n

var (
	cn = map[string]string{
		"all_email":             "全部邮件数据",
		"inbox":                 "收件箱",
		"outbox":                "发件箱",
		"sketch":                "草稿箱",
		"aperror":               "账号或密码错误",
		"unknowError":           "未知错误",
		"succ":                  "成功",
		"send_fail":             "发送失败",
		"att_err":               "附件解码错误",
		"login_exp":             "登录已失效",
		"ip_taps":               "这是你服务器IP，确保这个IP正确",
		"invalid_email_address": "无效的邮箱地址！",
		"deleted":               "已删除",
		"junk":                  "广告箱",
	}
	en = map[string]string{
		"all_email":             "All Email",
		"inbox":                 "Inbox",
		"outbox":                "Outbox",
		"sketch":                "Sketch",
		"aperror":               "Incorrect account number or password",
		"unknowError":           "Unknow Error",
		"succ":                  "Success",
		"send_fail":             "Send Failure",
		"att_err":               "Attachment decoding error",
		"login_exp":             "Login has expired.",
		"ip_taps":               "This is your server's IP, make sure it is correct.",
		"invalid_email_address": "Invalid e-mail address!",
		"deleted":               "Deleted",
		"junk":                  "Junk",
	}
)

func GetText(lang, key string) string {
	if lang == "zhCn" {
		text, exist := cn[key]
		if !exist {
			return ""
		}
		return text
	}
	text, exist := en[key]
	if !exist {
		return ""
	}
	return text
}
