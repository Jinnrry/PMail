package parsemail

import (
	"fmt"
	"pmail/config"
	"testing"
)

func TestEmail_domainMatch(t *testing.T) {
	//e := &Email{}
	//dnsNames := []string{
	//	"*.mail.qq.com",
	//	"993.dav.qq.com",
	//	"993.eas.qq.com",
	//	"993.imap.qq.com",
	//	"993.pop.qq.com",
	//	"993.smtp.qq.com",
	//	"imap.qq.com",
	//	"mx1.qq.com",
	//	"mx2.qq.com",
	//	"mx3.qq.com",
	//	"pop.qq.com",
	//	"smtp.qq.com",
	//	"mail.qq.com",
	//}
	//
	//fmt.Println(e.domainMatch("", dnsNames))
	//fmt.Println(e.domainMatch("xjiangwei.cn", dnsNames))
	//fmt.Println(e.domainMatch("qq.com", dnsNames))
	//fmt.Println(e.domainMatch("test.aaa.mail.qq.com", dnsNames))
	//fmt.Println(e.domainMatch("smtp.qq.com", dnsNames))
	//fmt.Println(e.domainMatch("pop.qq.com", dnsNames))
	//fmt.Println(e.domainMatch("test.mail.qq.com", dnsNames))

}

func Test_buildUser(t *testing.T) {
	u := buildUser("Jinnrry N <jiangwei1995910@gmail.com>")
	if u.EmailAddress != "jiangwei1995910@gmail.com" {
		t.Error("error")
	}
	if u.Name != "Jinnrry N" {
		t.Error("error")
	}
}

func TestEmail_BuilderHeaders(t *testing.T) {
	config.Init()
	Init()
	e := Email{
		From: buildUser("Jinnrry N <jiangwei1995910@gmail.com>"),
	}
	fmt.Println(string(e.BuilderHeaders(nil)))
}
