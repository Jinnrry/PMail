package parsemail

import (
	"bytes"
	"fmt"
	"github.com/emersion/go-message"
	"io"
	"testing"
)

func Test_buildUser(t *testing.T) {
	u := buildUser("Jinnrry N <jiangwei1995910@gmail.com>")
	if u.EmailAddress != "jiangwei1995910@gmail.com" {
		t.Error("error")
	}
	if u.Name != "Jinnrry N" {
		t.Error("error")
	}

	u = buildUser("=?UTF-8?B?YWRtaW5AamlubnJyeS5jb20=?=<admin@jinnrry.com>")
	if u.EmailAddress != "admin@jinnrry.com" {
		t.Error("error")
	}
	if u.Name != "admin@jinnrry.com" {
		t.Error("error")
	}

	u = buildUser("\"admin@jinnrry.com\" <admin@jinnrry.com>")
	if u.EmailAddress != "admin@jinnrry.com" {
		t.Error("error")
	}
	if u.Name != "admin@jinnrry.com" {
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
		Subject: "Title中文",
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
