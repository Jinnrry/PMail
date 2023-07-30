package email

import (
	"encoding/base64"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"pmail/config"
	"pmail/dto"
	"pmail/dto/parsemail"
	"pmail/dto/response"
	"pmail/hooks"
	"pmail/i18n"
	"pmail/mysql"
	"pmail/smtp_server"
	"pmail/utils/async"
	"strings"
	"time"
)

type sendRequest struct {
	ReplyTo     []user       `json:"reply_to"`
	From        user         `json:"from"`
	To          []user       `json:"to"`
	Bcc         []user       `json:"bcc"`
	Cc          []user       `json:"cc"`
	Subject     string       `json:"subject"`
	Text        string       `json:"text"`   // Plaintext message (optional)
	HTML        string       `json:"html"`   // Html message (optional)
	Sender      user         `json:"sender"` // override From as SMTP envelope sender (optional)
	ReadReceipt []string     `json:"read_receipt"`
	Attachments []attachment `json:"attrs"`
}

type user struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type attachment struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func Send(ctx *dto.Context, w http.ResponseWriter, req *http.Request) {
	reqBytes, err := io.ReadAll(req.Body)
	if err != nil {
		log.WithContext(ctx).Errorf("%+v", err)
		response.NewErrorResponse(response.ParamsError, "params error", err.Error()).FPrint(w)
		return
	}
	log.WithContext(ctx).Infof("发送邮件")

	var reqData sendRequest
	err = json.Unmarshal(reqBytes, &reqData)
	if err != nil {
		log.WithContext(ctx).Errorf("%+v", err)
		response.NewErrorResponse(response.ParamsError, "params error", err.Error()).FPrint(w)
		return
	}

	if reqData.From.Email == "" && reqData.From.Name != "" {
		reqData.From.Email = reqData.From.Name + "@" + config.Instance.Domain
	}

	if reqData.From.Email == "" {
		response.NewErrorResponse(response.ParamsError, "发件人必填", "发件人必填").FPrint(w)
		return
	}

	if reqData.Subject == "" {
		response.NewErrorResponse(response.ParamsError, "邮件标题必填", "邮件标题必填").FPrint(w)
		return
	}

	if len(reqData.To) <= 0 {
		response.NewErrorResponse(response.ParamsError, "收件人必填", "收件人必填").FPrint(w)
		return
	}

	e := &parsemail.Email{}

	for _, to := range reqData.To {
		e.To = append(e.To, &parsemail.User{
			Name:         to.Name,
			EmailAddress: to.Email,
		})
	}

	for _, bcc := range reqData.Bcc {
		e.Bcc = append(e.Bcc, &parsemail.User{
			Name:         bcc.Name,
			EmailAddress: bcc.Email,
		})
	}

	for _, cc := range reqData.Cc {
		e.Cc = append(e.Cc, &parsemail.User{
			Name:         cc.Name,
			EmailAddress: cc.Email,
		})
	}

	e.From = &parsemail.User{
		Name:         reqData.From.Name,
		EmailAddress: reqData.From.Email,
	}
	e.Text = []byte(reqData.Text)
	e.HTML = []byte(reqData.HTML)
	e.Subject = reqData.Subject
	for _, att := range reqData.Attachments {
		att.Data = strings.TrimPrefix(att.Data, "data:")
		infos := strings.Split(att.Data, ";")
		contentType := infos[0]
		content := strings.TrimPrefix(infos[1], "base64,")
		decoded, err := base64.StdEncoding.DecodeString(content)
		if err != nil {
			log.WithContext(ctx).Errorf("附件解码错误！%v", err)
			response.NewErrorResponse(response.ParamsError, i18n.GetText(ctx.Lang, "att_err"), err.Error()).FPrint(w)
			return
		}
		e.Attachments = append(e.Attachments, &parsemail.Attachment{
			Filename:    att.Name,
			ContentType: contentType,
			Content:     decoded,
		})

	}

	for _, hook := range hooks.HookList {
		if hook == nil {
			continue
		}
		async.New(ctx).Process(func() {
			hook.SendBefore(ctx, e)
		})
	}

	// 邮件落库
	sql := "INSERT INTO email (type,subject, reply_to, from_name, from_address, `to`, bcc, cc, text, html, sender, attachments,spf_check, dkim_check, create_time,send_user_id,error) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	sqlRes, sqlerr := mysql.Instance.Exec(mysql.WithContext(ctx, sql),
		1,
		e.Subject,
		json2string(e.ReplyTo),
		e.From.Name,
		e.From.EmailAddress,
		json2string(e.To),
		json2string(e.Bcc),
		json2string(e.Cc),
		e.Text,
		e.HTML,
		json2string(e.Sender),
		json2string(e.Attachments),
		1,
		1,
		time.Now(),
		ctx.UserInfo.ID,
		"",
	)
	emailId, _ := sqlRes.LastInsertId()

	if sqlerr != nil || emailId <= 0 {
		log.Println("mysql insert error:", err.Error())
		response.NewErrorResponse(response.ServerError, i18n.GetText(ctx.Lang, "send_fail"), err.Error()).FPrint(w)
		return
	}

	async.New(ctx).Process(func() {
		errMsg := ""
		err, sendErr := smtp_server.Send(ctx, e)

		for _, hook := range hooks.HookList {
			if hook == nil {
				continue
			}
			async.New(ctx).Process(func() {
				hook.SendAfter(ctx, e, sendErr)
			})
		}

		if err != nil {
			errMsg = err.Error()
			_, err := mysql.Instance.Exec(mysql.WithContext(ctx, "update email set status =2 ,error=? where id = ? "), errMsg, emailId)
			if err != nil {
				log.WithContext(ctx).Errorf("sql Error :%+v", err)
			}
		} else {
			_, err := mysql.Instance.Exec(mysql.WithContext(ctx, "update email set status =1  where id = ? "), emailId)
			if err != nil {
				log.WithContext(ctx).Errorf("sql Error :%+v", err)
			}
		}

	})

	response.NewSuccessResponse(i18n.GetText(ctx.Lang, "succ")).FPrint(w)
}

func json2string(d any) string {
	by, _ := json.Marshal(d)
	return string(by)
}
