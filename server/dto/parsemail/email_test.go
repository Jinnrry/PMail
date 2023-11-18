package parsemail

import (
	"bytes"
	"fmt"
	"github.com/emersion/go-message"
	"io"
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

func TestEmailBuidlers(t *testing.T) {
	var b bytes.Buffer

	var h message.Header
	h.SetContentType("multipart/alternative", nil)
	w, err := message.CreateWriter(&b, h)
	if err != nil {
	}

	var h1 message.Header
	h1.SetContentType("text/html", nil)
	w1, err := w.CreatePart(h1)
	if err != nil {
	}
	io.WriteString(w1, "<h1>Hello World!</h1><p>This is an HTML part.</p>")
	w1.Close()

	var h2 message.Header
	h2.SetContentType("text/plain", nil)
	w2, err := w.CreatePart(h2)
	if err != nil {
	}
	io.WriteString(w2, "Hello World!\n\nThis is a text part.")
	w2.Close()

	w.Close()

	fmt.Println(b.String())
}

func TestEmail_builder(t *testing.T) {
	e := Email{
		From:    buildUser("i@test.com"),
		To:      buildUsers([]string{"to@test.com"}),
		Subject: "Title",
		HTML:    []byte("Html"),
		Text:    []byte("Text"),
		Attachments: []*Attachment{
			{
				Filename:    "a.png",
				ContentType: "image/jpeg",
				Content:     []byte("aaa"),
				ContentID:   "1",
			},
		},
	}

	rest := e.BuildBytes(nil, false)
	fmt.Println(string(rest))
}
