package parsemail

import (
	"bytes"
	"github.com/emersion/go-message"
	_ "github.com/emersion/go-message/charset"
	"github.com/emersion/go-message/mail"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"io"
	"net/textproto"
	"pmail/utils/array"
	"pmail/utils/context"
	"regexp"
	"strings"
	"time"
)

type User struct {
	EmailAddress string `json:"EmailAddress"`
	Name         string `json:"Name"`
}

func (u User) GetDomainAccount() (string, string) {
	infos := strings.Split(u.EmailAddress, "@")
	if len(infos) >= 2 {
		return infos[0], infos[1]
	}

	return "", ""
}

type Attachment struct {
	Filename    string
	ContentType string
	Content     []byte
	ContentID   string
}

// Email is the type used for email messages
type Email struct {
	ReplyTo     []*User
	From        *User
	To          []*User
	Bcc         []*User
	Cc          []*User
	Subject     string
	Text        []byte // Plaintext message (optional)
	HTML        []byte // Html message (optional)
	Sender      *User  // override From as SMTP envelope sender (optional)
	Headers     textproto.MIMEHeader
	Attachments []*Attachment
	ReadReceipt []string
	Date        string
	IsRead      int
	Status      int // 0未发送，1已发送，2发送失败，3删除
	GroupId     int // 分组id
	MessageId   int64
}

func NewEmailFromReader(to []string, r io.Reader) *Email {
	ret := &Email{}
	m, err := message.Read(r)
	if err != nil {
		log.Errorf("email解析错误！ Error %+v", err)
	}

	ret.From = buildUser(m.Header.Get("From"))

	if len(to) > 0 {
		ret.To = buildUsers(to)
	} else {
		ret.To = buildUsers(m.Header.Values("To"))
	}

	ret.Cc = buildUsers(m.Header.Values("Cc"))
	ret.ReplyTo = buildUsers(m.Header.Values("ReplyTo"))
	ret.Sender = buildUser(m.Header.Get("Sender"))
	if ret.Sender == nil {
		ret.Sender = ret.From
	}

	ret.Subject, _ = m.Header.Text("Subject")

	sendTime, err := time.Parse(time.RFC1123Z, m.Header.Get("Date"))
	if err != nil {
		sendTime = time.Now()
	}
	ret.Date = sendTime.Format(time.DateTime)
	m.Walk(func(path []int, entity *message.Entity, err error) error {
		return formatContent(entity, ret)
	})
	return ret
}

func formatContent(entity *message.Entity, ret *Email) error {
	contentType, p, err := entity.Header.ContentType()

	if err != nil {
		log.Errorf("email read error! %+v", err)
		return err
	}

	switch contentType {
	case "multipart/alternative":
	case "multipart/mixed":
	case "text/plain":
		ret.Text, _ = io.ReadAll(entity.Body)
	case "text/html":
		ret.HTML, _ = io.ReadAll(entity.Body)
	case "multipart/related":
		entity.Walk(func(path []int, entity *message.Entity, err error) error {
			if t, _, _ := entity.Header.ContentType(); t == "multipart/related" {
				return nil
			}
			return formatContent(entity, ret)
		})
	default:
		c, _ := io.ReadAll(entity.Body)
		fileName := p["name"]
		if fileName == "" {
			contentDisposition := entity.Header.Get("Content-Disposition")
			r := regexp.MustCompile("filename=(.*)")
			matchs := r.FindStringSubmatch(contentDisposition)
			if len(matchs) == 2 {
				fileName = matchs[1]
			} else {
				fileName = "no_name_file"
			}
		}

		ret.Attachments = append(ret.Attachments, &Attachment{
			Filename:    fileName,
			ContentType: contentType,
			Content:     c,
			ContentID:   strings.TrimPrefix(strings.TrimSuffix(entity.Header.Get("Content-Id"), ">"), "<"),
		})
	}

	return nil
}

func BuilderUser(str string) *User {
	return buildUser(str)
}

func buildUser(str string) *User {
	if str == "" {
		return nil
	}

	ret := &User{}
	args := strings.Split(str, " ")
	if len(args) == 1 {
		ret.EmailAddress = str
		return ret
	}

	if len(args) > 2 {
		targs := []string{
			array.Join(args[0:len(args)-1], " "),
			args[len(args)-1],
		}
		args = targs
	}

	args[0] = strings.Trim(args[0], "\"")
	args[1] = strings.TrimPrefix(args[1], "<")
	args[1] = strings.TrimSuffix(args[1], ">")

	name, err := (&WordDecoder{}).Decode(strings.ReplaceAll(args[0], "\"", ""))
	if err == nil {
		ret.Name = name
	} else {
		ret.Name = args[0]
	}
	ret.EmailAddress = args[1]
	return ret
}

func buildUsers(str []string) []*User {
	var ret []*User
	for _, s1 := range str {
		for _, s := range strings.Split(s1, ",") {
			s = strings.TrimSpace(s)
			ret = append(ret, buildUser(s))
		}
	}

	return ret
}

func (e *Email) ForwardBuildBytes(ctx *context.Context, forwardAddress string) []byte {
	var b bytes.Buffer

	from := []*mail.Address{{e.From.Name, e.From.EmailAddress}}
	to := []*mail.Address{
		{
			Address: forwardAddress,
		},
	}

	// Create our mail header
	var h mail.Header
	h.SetDate(time.Now())
	h.SetAddressList("From", from)
	h.SetAddressList("To", to)
	h.SetText("Subject", e.Subject)
	h.SetMessageID(cast.ToString(e.MessageId))
	if len(e.Cc) != 0 {
		cc := []*mail.Address{}
		for _, user := range e.Cc {
			cc = append(cc, &mail.Address{
				Name:    user.Name,
				Address: user.EmailAddress,
			})
		}
		h.SetAddressList("Cc", cc)
	}

	// Create a new mail writer
	mw, err := mail.CreateWriter(&b, h)
	if err != nil {
		log.WithContext(ctx).Fatal(err)
	}

	// Create a text part
	tw, err := mw.CreateInline()
	if err != nil {
		log.WithContext(ctx).Fatal(err)
	}
	var th mail.InlineHeader
	th.Set("Content-Type", "text/plain")
	w, err := tw.CreatePart(th)
	if err != nil {
		log.Fatal(err)
	}
	io.WriteString(w, string(e.Text))
	w.Close()

	var html mail.InlineHeader
	html.Set("Content-Type", "text/html")
	w, err = tw.CreatePart(html)
	if err != nil {
		log.Fatal(err)
	}
	io.WriteString(w, string(e.HTML))
	w.Close()

	tw.Close()

	// Create an attachment
	for _, attachment := range e.Attachments {
		var ah mail.AttachmentHeader
		ah.Set("Content-Type", attachment.ContentType)
		ah.SetFilename(attachment.Filename)
		w, err = mw.CreateAttachment(ah)
		if err != nil {
			log.WithContext(ctx).Fatal(err)
			continue
		}
		w.Write(attachment.Content)
		w.Close()
	}

	mw.Close()

	// dkim 签名后返回
	return instance.Sign(b.String())
}

func (e *Email) BuildBytes(ctx *context.Context, dkim bool) []byte {
	var b bytes.Buffer

	from := []*mail.Address{{e.From.Name, e.From.EmailAddress}}
	to := []*mail.Address{}
	for _, user := range e.To {
		to = append(to, &mail.Address{
			Name:    user.Name,
			Address: user.EmailAddress,
		})
	}

	// Create our mail header
	var h mail.Header
	if e.Date != "" {
		t, err := time.ParseInLocation("2006-01-02 15:04:05", e.Date, time.Local)
		if err != nil {
			log.WithContext(ctx).Errorf("Time Error ! Err:%+v", err)
			h.SetDate(time.Now())
		} else {
			h.SetDate(t)
		}
	} else {
		h.SetDate(time.Now())
	}
	h.SetMessageID(cast.ToString(e.MessageId))
	h.SetAddressList("From", from)
	h.SetAddressList("To", to)
	h.SetText("Subject", e.Subject)
	if len(e.Cc) != 0 {
		cc := []*mail.Address{}
		for _, user := range e.Cc {
			cc = append(cc, &mail.Address{
				Name:    user.Name,
				Address: user.EmailAddress,
			})
		}
		h.SetAddressList("Cc", cc)
	}

	// Create a new mail writer
	mw, err := mail.CreateWriter(&b, h)
	if err != nil {
		log.WithContext(ctx).Fatal(err)
	}

	// Create a text part
	tw, err := mw.CreateInline()
	if err != nil {
		log.WithContext(ctx).Fatal(err)
	}
	var th mail.InlineHeader
	th.SetContentType("text/plain", map[string]string{
		"charset": "UTF-8",
	})
	w, err := tw.CreatePart(th)
	if err != nil {
		log.Fatal(err)
	}
	io.WriteString(w, string(e.Text))
	w.Close()

	var html mail.InlineHeader
	html.SetContentType("text/html", map[string]string{
		"charset": "UTF-8",
	})
	w, err = tw.CreatePart(html)
	if err != nil {
		log.Fatal(err)
	}
	if len(e.HTML) > 0 {
		io.WriteString(w, string(e.HTML))
	} else {
		io.WriteString(w, string(e.Text))
	}

	w.Close()

	tw.Close()

	// Create an attachment
	for _, attachment := range e.Attachments {
		var ah mail.AttachmentHeader
		ah.Set("Content-Type", attachment.ContentType)
		ah.SetFilename(attachment.Filename)
		w, err = mw.CreateAttachment(ah)
		if err != nil {
			log.WithContext(ctx).Fatal(err)
			continue
		}
		w.Write(attachment.Content)
		w.Close()
	}

	mw.Close()

	if dkim {
		// dkim 签名后返回
		return instance.Sign(b.String())
	}
	return b.Bytes()
}
