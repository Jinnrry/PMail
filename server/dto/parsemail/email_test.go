package parsemail

import (
	"bytes"
	"fmt"
	"io"
	stdmail "net/mail"
	"strings"
	"testing"
	"time"

	"github.com/Jinnrry/pmail/models"
	"github.com/emersion/go-message"
)

func builtMessageDate(t *testing.T, raw []byte) time.Time {
	t.Helper()

	m, err := stdmail.ReadMessage(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("read built message: %v", err)
	}
	date, err := stdmail.ParseDate(m.Header.Get("Date"))
	if err != nil {
		t.Fatalf("parse built Date header %q: %v", m.Header.Get("Date"), err)
	}
	return date
}

func TestEmailDateFromReaderPreservesInstant(t *testing.T) {
	const dateHeader = "Sun, 09 Aug 2026 11:51:08 +0000"
	raw := []byte("From: sender@example.com\r\n" +
		"To: recipient@example.com\r\n" +
		"Subject: Date round trip\r\n" +
		"Date: " + dateHeader + "\r\n" +
		"Message-ID: <date-reader@example.com>\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n\r\n" +
		"body")

	expected, err := stdmail.ParseDate(dateHeader)
	if err != nil {
		t.Fatalf("parse fixture date: %v", err)
	}

	email := NewEmailFromReader([]string{"recipient@example.com"}, bytes.NewReader(raw), len(raw))
	stored, err := time.Parse(time.RFC3339, email.Date)
	if err != nil {
		t.Fatalf("stored Date is not RFC3339: %q: %v", email.Date, err)
	}
	if !stored.Equal(expected) {
		t.Fatalf("stored Date changed instant: got %s, want %s", stored, expected)
	}

	got := builtMessageDate(t, email.BuildBytes(nil, false))
	if !got.Equal(expected) {
		t.Fatalf("rebuilt Date changed instant: got %s, want %s", got, expected)
	}
}

func TestEmailDateFromModelPreservesUTCInstant(t *testing.T) {
	expected := time.Date(2026, time.August, 9, 11, 51, 8, 0, time.UTC)
	email := NewEmailFromModel(models.Email{
		Id:          1,
		FromAddress: "sender@example.com",
		To:          `[{"EmailAddress":"recipient@example.com","Name":""}]`,
		Subject:     "Model date",
		SendDate:    expected,
		MsgID:       "date-model@example.com",
	})

	if email.Date != expected.Format(time.RFC3339) {
		t.Fatalf("model Date = %q, want %q", email.Date, expected.Format(time.RFC3339))
	}

	got := builtMessageDate(t, email.BuildBytes(nil, false))
	if !got.Equal(expected) {
		t.Fatalf("rebuilt model Date changed instant: got %s, want %s", got, expected)
	}
}

func TestEmailBuildBytesAcceptsLegacyDate(t *testing.T) {
	const legacyDate = "2026-08-09 11:51:08"
	expected, err := time.ParseInLocation(time.DateTime, legacyDate, time.Local)
	if err != nil {
		t.Fatalf("parse legacy fixture date: %v", err)
	}

	email := &Email{
		From:    buildUser("sender@example.com"),
		To:      buildUsers([]string{"recipient@example.com"}),
		Subject: "Legacy date",
		Text:    []byte("body"),
		Date:    legacyDate,
		MsgID:   "date-legacy@example.com",
	}

	got := builtMessageDate(t, email.BuildBytes(nil, false))
	if !got.Equal(expected) {
		t.Fatalf("legacy Date changed instant: got %s, want %s", got, expected)
	}
}

func TestHtmlTxtAttachment(t *testing.T) {
	emailBytes := `From: "=?utf-8?B?amlubnJyeQ==?=" <ok@xjiangwei.cn>
To: "=?utf-8?B?YWRtaW4=?=" <admin@jinnrry.com>
Subject: FileTest
Mime-Version: 1.0
Content-Type: multipart/mixed;
        boundary="----=_NextPart_6A102A6A_F16B7280_56878F2C"
Content-Transfer-Encoding: 8Bit
Date: Fri, 22 May 2026 18:05:30 +0800
Message-ID: <tencent_CB10C599AF69D91666D5A66E4505F633A705@qq.com>

This is a multi-part message in MIME format.

------=_NextPart_6A102A6A_F16B7280_56878F2C
Content-Type: multipart/alternative;
        boundary="----=_NextPart_6A102A6A_F16B7280_0A1A2221";

------=_NextPart_6A102A6A_F16B7280_0A1A2221
Content-Type: text/plain;
        charset="utf-8"
Content-Transfer-Encoding: base64

RW1haWwgQ29udGVudCE=

------=_NextPart_6A102A6A_F16B7280_0A1A2221
Content-Type: text/html;
        charset="utf-8"
Content-Transfer-Encoding: base64

PGRpdiBzdHlsZT0iZm9udC1mYW1pbHk6IC1hcHBsZS1zeXN0ZW0sIHN5c3RlbS11aTsgZm9u
dC1zaXplOiAxNHB4OyBjb2xvcjogcmdiKDAsIDAsIDApOyBsaW5lLWhlaWdodDogMS40Mzsi
PjxiPkVtYWlsIENvbnRlbnQhPC9iPjwvZGl2Pg==

------=_NextPart_6A102A6A_F16B7280_0A1A2221--

------=_NextPart_6A102A6A_F16B7280_56878F2C
Content-Type: application/octet-stream;
        charset="utf-8";
        name="html.html"
Content-Disposition: attachment; filename="html.html"
Content-Transfer-Encoding: base64

dGhpcyBpcyBodG1sIGZpbGUgY29udGV4dCE=

------=_NextPart_6A102A6A_F16B7280_56878F2C
Content-Type: application/octet-stream;
        charset="utf-8";
        name="txt.txt"
Content-Disposition: attachment; filename="txt.txt"
Content-Transfer-Encoding: base64

dGhpcyBpcyB0eHQgZmlsZSBjb250ZW50IQ==

------=_NextPart_6A102A6A_F16B7280_56878F2C--`

	email := NewEmailFromReader([]string{"admin@jinnrry.com"}, bytes.NewReader([]byte(emailBytes)), len(emailBytes))

	if strings.Contains(string(email.Text), "file") {
		t.Errorf("formatContentError:%s", "邮件内容解析错误，文件内容被解析到邮件中！")
	}

	if strings.Contains(string(email.HTML), "file") {
		t.Errorf("formatContentError:%s", "邮件内容解析错误，文件内容被解析到邮件中！")
	}

	emailBytes = `Content-Type: multipart/mixed;
 boundary=ee32d21f4234ae8884badf33c4eec57d61a93bc5048453d42dd9313d5974
Mime-Version: 1.0
Subject: TestHtmlTxtAttachment
To: <admin@jinnrry.com>
Sender: "jinnrry" <admin@jinnrry.com>
From: "jinnrry" <admin@jinnrry.com>
Message-Id: <0f376803f71ff716c49bdf0479c5568c.1779445057831143578@jinnrry.com>
Date: Fri, 22 May 2026 18:17:37 +0800

--ee32d21f4234ae8884badf33c4eec57d61a93bc5048453d42dd9313d5974
Content-Type: multipart/alternative;
 boundary=71b61b51b82589aba3a3315cc61b6406a790a8d6e50c34782a4f2149134d

--71b61b51b82589aba3a3315cc61b6406a790a8d6e50c34782a4f2149134d
Content-Disposition: inline
Content-Type: text/plain; charset=UTF-8
Content-Transfer-Encoding: base64

SGVyZSBpcyBlbWFpbCBjb250ZW50Lg==
--71b61b51b82589aba3a3315cc61b6406a790a8d6e50c34782a4f2149134d
Content-Disposition: inline
Content-Transfer-Encoding: base64
Content-Type: text/html; charset=UTF-8

PHA+PHN0cm9uZz5IZXJlIGlzIGVtYWlsIGNvbnRlbnQuPC9zdHJvbmc+PC9wPg==
--71b61b51b82589aba3a3315cc61b6406a790a8d6e50c34782a4f2149134d--

--ee32d21f4234ae8884badf33c4eec57d61a93bc5048453d42dd9313d5974
Content-Transfer-Encoding: base64
Content-Disposition: attachment; filename=html.html
Content-Type: text/html

dGhpcyBpcyBodG1sIGZpbGUgY29udGV4dCE=
--ee32d21f4234ae8884badf33c4eec57d61a93bc5048453d42dd9313d5974
Content-Transfer-Encoding: base64
Content-Disposition: attachment; filename=txt.txt
Content-Type: text/plain

dGhpcyBpcyB0eHQgZmlsZSBjb250ZW50IQ==
--ee32d21f4234ae8884badf33c4eec57d61a93bc5048453d42dd9313d5974--`

	email = NewEmailFromReader([]string{"admin@jinnrry.com"}, bytes.NewReader([]byte(emailBytes)), len(emailBytes))

	if strings.Contains(string(email.Text), "file") {
		t.Errorf("formatContentError:%s", "邮件内容解析错误，文件内容被解析到邮件中！")
	}

	if strings.Contains(string(email.HTML), "file") {
		t.Errorf("formatContentError:%s", "邮件内容解析错误，文件内容被解析到邮件中！")
	}

}

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

func TestEmail_BuildPart(t *testing.T) {
	e := Email{
		Text: []byte("text"),
		HTML: []byte("html"),
	}
	res := e.BuildPart(nil, []int{1, 2})
	fmt.Println(string(res))

}
