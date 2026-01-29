package parsemail

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/textproto"
	"regexp"
	"strings"
	"time"

	"github.com/Jinnrry/pmail/config"
	"github.com/Jinnrry/pmail/models"
	"github.com/Jinnrry/pmail/utils/context"
	"github.com/emersion/go-message"
	_ "github.com/emersion/go-message/charset"
	"github.com/emersion/go-message/mail"
	"github.com/microcosm-cc/bluemonday"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
)

type User struct {
	EmailAddress string `json:"EmailAddress"`
	Name         string `json:"Name"`
}

func (u User) Build() string {
	if u.Name != "" {
		return fmt.Sprintf("\"%s\" <%s>", mime.QEncoding.Encode("utf-8", u.Name), u.EmailAddress)
	}
	return fmt.Sprintf("<%s>", u.EmailAddress)
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
	Status      int // 0未发送，1已发送，2发送失败，3删除，5广告邮件
	MessageId   int64
	Size        int
}

// Xss filter policy
var (
	strictPolicy  *bluemonday.Policy
	relaxedPolicy *bluemonday.Policy
)

func init() {
	strictPolicy = bluemonday.StrictPolicy()

	relaxedPolicy = bluemonday.NewPolicy()

	relaxedPolicy.AllowElements("p", "br", "strong", "em", "u", "b", "i", "h1", "h2", "h3", "h4", "h5", "h6")
	relaxedPolicy.AllowElements("div", "span", "center")
	relaxedPolicy.AllowElements("ul", "ol", "li")
	relaxedPolicy.AllowElements("blockquote", "cite")

	relaxedPolicy.AllowElements("table", "tbody", "thead", "tr", "td", "th")
	relaxedPolicy.AllowAttrs("width", "height", "border", "cellpadding", "cellspacing").OnElements("table")
	relaxedPolicy.AllowAttrs("align", "valign", "colspan", "rowspan").OnElements("td", "th")
	relaxedPolicy.AllowAttrs("align").OnElements("tr")

	relaxedPolicy.AllowAttrs("style").Globally()
	relaxedPolicy.AllowAttrs("class", "id").Globally()

	relaxedPolicy.AllowAttrs("bgcolor", "color", "background").Globally()
	relaxedPolicy.AllowAttrs("align").OnElements("p", "div", "h1", "h2", "h3", "h4", "h5", "h6")

	relaxedPolicy.AllowElements("img")
	relaxedPolicy.AllowAttrs("src", "alt", "width", "height", "style", "align").OnElements("img")

	relaxedPolicy.AllowElements("a")
	relaxedPolicy.AllowAttrs("href", "style").OnElements("a")
	relaxedPolicy.RequireNoReferrerOnLinks(true)
	relaxedPolicy.AddTargetBlankToFullyQualifiedLinks(true)
	relaxedPolicy.RequireNoFollowOnLinks(true)

	relaxedPolicy.AllowElements("font")
	relaxedPolicy.AllowAttrs("size", "color", "face").OnElements("font")

	relaxedPolicy.AllowElements("style")
	relaxedPolicy.AllowAttrs("type").OnElements("style")

	relaxedPolicy.AllowURLSchemes("http", "https", "mailto")

	relaxedPolicy.SkipElementsContent("script", "object", "embed", "iframe", "frame", "frameset")
}

func sanitizeHTML(htmlContent string) string {
	if htmlContent == "" {
		return ""
	}

	sanitized := relaxedPolicy.Sanitize(htmlContent)

	dataUrlRegex := regexp.MustCompile(`href\s*=\s*["']data:[^"']*["']`)
	sanitized = dataUrlRegex.ReplaceAllString(sanitized, `rel="nofollow"`)

	jsUrlRegex := regexp.MustCompile(`href\s*=\s*["']javascript:[^"']*["']`)
	sanitized = jsUrlRegex.ReplaceAllString(sanitized, `rel="nofollow"`)

	expressionRegex := regexp.MustCompile(`(?i)expression\s*\(.*?\)`)
	sanitized = expressionRegex.ReplaceAllString(sanitized, "")

	styleExpressionRegex := regexp.MustCompile(`(?i)style\s*=\s*["'][^"']*expression[^"']*["']`)
	sanitized = styleExpressionRegex.ReplaceAllString(sanitized, "")

	cssJsRegex := regexp.MustCompile(`(?i)javascript\s*:`)
	sanitized = cssJsRegex.ReplaceAllString(sanitized, "")

	return sanitized
}

// Sanitize Text
func sanitizeText(text string) string {
	return strictPolicy.Sanitize(text)
}

func users2String(users []*User) string {
	ret := ""
	for _, user := range users {
		if ret != "" {
			ret += ", "
		}
		ret += user.Build()
	}
	return ret
}

func (e *Email) BuildTo2String() string {
	return users2String(e.To)
}

func (e *Email) BuildCc2String() string {
	return users2String(e.Cc)
}

func (e *Email) BuildBcc2String() string {
	return users2String(e.Bcc)
}

func NewEmailFromModel(d models.Email) *Email {

	var To []*User
	json.Unmarshal([]byte(d.To), &To)

	var ReplyTo []*User
	json.Unmarshal([]byte(d.ReplyTo), &ReplyTo)

	var Sender *User
	json.Unmarshal([]byte(d.Sender), &Sender)

	var Bcc []*User
	json.Unmarshal([]byte(d.Bcc), &Bcc)

	var Cc []*User
	json.Unmarshal([]byte(d.Cc), &Cc)

	var Attachments []*Attachment
	json.Unmarshal([]byte(d.Attachments), &Attachments)

	return &Email{
		MessageId: cast.ToInt64(d.Id),
		From: &User{
			Name:         d.FromName,
			EmailAddress: d.FromAddress,
		},
		To:          To,
		Subject:     d.Subject,
		Text:        []byte(d.Text.String),
		HTML:        []byte(d.Html.String),
		Sender:      Sender,
		ReplyTo:     ReplyTo,
		Bcc:         Bcc,
		Cc:          Cc,
		Attachments: Attachments,
		Date:        d.SendDate.Format("2006-01-02 15:04:05"),
	}
}

func NewEmailFromReader(to []string, r io.Reader, size int) *Email {
	ret := &Email{}
	m, err := message.Read(r)
	if err != nil {
		log.Errorf("email解析错误！ Error %+v", err)
	}

	ret.Size = size
	ret.From = buildUser(m.Header.Get("From"))

	smtpTo := buildUsers(to)

	ret.To = buildUsers(m.Header.Values("To"))

	ret.Bcc = []*User{}

	for _, user := range smtpTo {
		in := false
		for _, u := range ret.To {
			if u.EmailAddress == user.EmailAddress {
				in = true
				break
			}
		}
		if !in {
			ret.Bcc = append(ret.Bcc, user)
		}

	}

	ret.Cc = buildUsers(m.Header.Values("Cc"))
	ret.ReplyTo = buildUsers(m.Header.Values("ReplyTo"))
	ret.Sender = buildUser(m.Header.Get("Sender"))
	if ret.Sender == nil {
		ret.Sender = ret.From
	}

	subject, _ := m.Header.Text("Subject")
	ret.Subject = strictPolicy.Sanitize(subject)

	sendTime, err := time.Parse(time.RFC1123Z, m.Header.Get("Date"))
	if err != nil {
		sendTime = time.Now()
	}
	ret.Date = sendTime.Format(time.DateTime)
	m.Walk(func(path []int, entity *message.Entity, err error) error {
		return formatContent(entity, ret)
	})

	if ret.From != nil {
		ret.From.Name = strictPolicy.Sanitize(ret.From.Name)
		ret.From.EmailAddress = strictPolicy.Sanitize(ret.From.EmailAddress)
	}

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
		testContent, _ := io.ReadAll(entity.Body)
		ret.Text = []byte(strictPolicy.Sanitize(string(testContent)))
	case "text/html":
		htmlContent, _ := io.ReadAll(entity.Body)
		ret.HTML = []byte(relaxedPolicy.Sanitize(string(htmlContent)))
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
			filenameRegex := regexp.MustCompile(`filename\s*=\s*"?([^";]+)"?`)
			matches := filenameRegex.FindStringSubmatch(contentDisposition)
			if len(matches) >= 2 {
				fileName = strings.TrimSpace(matches[1])
				fileName = strings.Trim(fileName, `"`)
			} else {
				fileName = "no_name_file"
			}
		}

		ret.Attachments = append(ret.Attachments, &Attachment{
			Filename:    sanitizeText(fileName),
			ContentType: sanitizeText(strings.TrimSpace(contentType)),
			Content:     c,
			ContentID:   strings.TrimPrefix(strings.TrimSuffix(entity.Header.Get("Content-Id"), ">"), "<"),
		})
	}

	return nil
}

func BuilderUser(str string) *User {
	return buildUser(str)
}

var emailAddressRe = regexp.MustCompile(`<(.*@.*)>`)

func buildUser(str string) *User {
	str = strings.TrimSpace(str)
	if str == "" {
		return &User{}
	}

	user := &User{}

	addr, err := mail.ParseAddress(str)
	if err == nil {
		user.EmailAddress = strings.TrimSpace(addr.Address)

		name := strings.TrimSpace(addr.Name)
		if name != "" {
			decoder := mime.WordDecoder{}
			if decoded, err := decoder.Decode(name); err == nil {
				name = decoded
			}
			user.Name = strictPolicy.Sanitize(name)
		}
		return user
	}

	matched := emailAddressRe.FindStringSubmatch(str)
	if len(matched) == 2 {
		user.EmailAddress = strings.TrimSpace(matched[1])
		namePart := strings.ReplaceAll(str, matched[0], "")
		namePart = strings.Trim(strings.TrimSpace(namePart), "\"")

		decoder := mime.WordDecoder{}
		if decoded, err := decoder.Decode(strings.ReplaceAll(namePart, "\"", "")); err == nil {
			user.Name = strictPolicy.Sanitize(strings.TrimSpace(decoded))
		} else {
			user.Name = strictPolicy.Sanitize(strings.TrimSpace(namePart))
		}
	} else {
		user.EmailAddress = strictPolicy.Sanitize(str)
	}

	return user
}

func buildUsers(strs []string) []*User {
	var ret []*User
	for _, line := range strs {
		if strings.TrimSpace(line) == "" {
			continue
		}

		parts := strings.Split(line, ",")
		for _, part := range parts {
			if u := buildUser(strings.TrimSpace(part)); u != nil {
				ret = append(ret, u)
			}
		}
	}
	return ret
}

func (e *Email) ForwardBuildBytes(ctx *context.Context, sender *models.User) []byte {
	var b bytes.Buffer

	from := []*mail.Address{{e.From.Name, e.From.EmailAddress}}
	to := []*mail.Address{}
	for _, user := range e.To {
		to = append(to, &mail.Address{
			Name:    user.Name,
			Address: user.EmailAddress,
		})
	}

	senderAddress := []*mail.Address{{sender.Name, fmt.Sprintf("%s@%s", sender.Account, config.Instance.Domains[0])}}
	// Create our mail header
	var h mail.Header
	h.SetDate(time.Now())
	h.SetAddressList("From", from)
	h.SetAddressList("Sender", senderAddress)
	h.SetAddressList("To", to)
	h.SetText("Subject", e.Subject)
	h.SetMessageID(fmt.Sprintf("%d@%s", e.MessageId, config.Instance.Domain))
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

func (e *Email) BuildPart(ctx *context.Context, loc []int) []byte {
	if len(loc) == 0 {
		return nil
	}

	// 处理顶层 part (part 1 = alternative, part 2+ = attachments)
	if len(loc) == 1 {
		partIdx := loc[0]
		if partIdx == 1 {
			// Part 1 是 alternative，不能直接获取，需要获取子部分
			return nil
		}
		// Part 2, 3, ... 是附件
		attachIdx := partIdx - 2
		if attachIdx >= 0 && attachIdx < len(e.Attachments) {
			encoded := base64.StdEncoding.EncodeToString(e.Attachments[attachIdx].Content)
			encoded += "\r\n"
			return []byte(encoded)
		}
		return nil
	}

	// 处理 alternative 的子部分 (1.1, 1.2)
	if loc[0] == 1 && len(loc) >= 2 {
		subIdx := loc[1]

		// 根据 BODYSTRUCTURE 的构建顺序：先 text，后 html
		// 如果只有一个存在，那个就是 1.1
		hasText := len(e.Text) > 0
		hasHtml := len(e.HTML) > 0

		if hasText && hasHtml {
			// 两者都有: 1.1=text, 1.2=html
			if subIdx == 1 {
				encoded := base64.StdEncoding.EncodeToString(e.Text)
				encoded += "\r\n"
				return []byte(encoded)
			}
			if subIdx == 2 {
				encoded := base64.StdEncoding.EncodeToString(e.HTML)
				encoded += "\r\n"
				return []byte(encoded)
			}
		} else if hasText {
			// 只有 text: 1.1=text
			if subIdx == 1 {
				encoded := base64.StdEncoding.EncodeToString(e.Text)
				encoded += "\r\n"
				return []byte(encoded)
			}
		} else if hasHtml {
			// 只有 html: 1.1=html
			if subIdx == 1 {
				encoded := base64.StdEncoding.EncodeToString(e.HTML)
				encoded += "\r\n"
				return []byte(encoded)
			}
		}
	}

	return nil
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
	h.SetMessageID(fmt.Sprintf("%d@%s", e.MessageId, config.Instance.Domain))
	h.SetAddressList("From", from)
	h.SetAddressList("Sender", from)
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

	if len(e.Text) > 0 {
		var th mail.InlineHeader
		th.Header.Set("Content-Transfer-Encoding", "base64")
		th.SetContentType("text/plain", map[string]string{
			"charset": "UTF-8",
		})
		w, err := tw.CreatePart(th)
		if err != nil {
			log.Fatal(err)
		}
		io.WriteString(w, string(e.Text))
		w.Close()
	}

	var html mail.InlineHeader
	html.SetContentType("text/html", map[string]string{
		"charset": "UTF-8",
	})
	html.Header.Set("Content-Transfer-Encoding", "base64")
	w, err := tw.CreatePart(html)
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
