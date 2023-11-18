package models

import (
	"database/sql"
	"encoding/json"
	"pmail/dto/parsemail"
	"time"
)

type Email struct {
	Id           int            `db:"id" json:"id"`
	Type         int8           `db:"type" json:"type"`
	GroupId      int            `db:"group_id" json:"group_id"`
	Subject      string         `db:"subject" json:"subject"`
	ReplyTo      string         `db:"reply_to" json:"reply_to"`
	FromName     string         `db:"from_name" json:"from_name"`
	FromAddress  string         `db:"from_address" json:"from_address"`
	To           string         `db:"to" json:"to"`
	Bcc          string         `db:"bcc" json:"bcc"`
	Cc           string         `db:"cc" json:"cc"`
	Text         sql.NullString `db:"text" json:"text"`
	Html         sql.NullString `db:"html" json:"html"`
	Sender       string         `db:"sender" json:"sender"`
	Attachments  string         `db:"attachments" json:"attachments"`
	SPFCheck     int8           `db:"spf_check" json:"spf_check"`
	DKIMCheck    int8           `db:"dkim_check" json:"dkim_check"`
	Status       int8           `db:"status" json:"status"` // 0未发送，1已发送，2发送失败，3删除
	CronSendTime time.Time      `db:"cron_send_time" json:"cron_send_time"`
	UpdateTime   time.Time      `db:"update_time" json:"update_time"`
	SendUserID   int            `db:"send_user_id" json:"send_user_id"`
	IsRead       int8           `db:"is_read" json:"is_read"`
	Error        sql.NullString `db:"error" json:"error"`
	SendDate     time.Time      `db:"send_date" json:"send_date"`
	CreateTime   time.Time      `db:"create_time" json:"create_time"`
}

type attachments struct {
	Filename    string
	ContentType string
	Index       int
	//Content     []byte
}

func (d Email) GetTos() []*parsemail.User {
	var ret []*parsemail.User
	json.Unmarshal([]byte(d.To), &ret)
	return ret
}

func (d Email) GetReplyTo() []*parsemail.User {
	var ret []*parsemail.User
	json.Unmarshal([]byte(d.ReplyTo), &ret)
	return ret
}

func (d Email) GetSender() *parsemail.User {
	var ret *parsemail.User
	json.Unmarshal([]byte(d.Sender), &ret)
	return ret
}

func (d Email) GetBcc() []*parsemail.User {
	var ret []*parsemail.User
	json.Unmarshal([]byte(d.Bcc), &ret)
	return ret
}

func (d Email) GetCc() []*parsemail.User {
	var ret []*parsemail.User
	json.Unmarshal([]byte(d.Cc), &ret)
	return ret
}

func (d Email) GetAttachments() []*parsemail.Attachment {
	var ret []*parsemail.Attachment
	json.Unmarshal([]byte(d.Attachments), &ret)
	return ret
}

func (d Email) MarshalJSON() ([]byte, error) {
	type Alias Email

	var allAtt = []attachments{}
	var showAtt = []attachments{}
	if d.Attachments != "" {
		_ = json.Unmarshal([]byte(d.Attachments), &allAtt)
		for i, att := range allAtt {
			att.Index = i
			if att.ContentType == "application/octet-stream" {
				showAtt = append(showAtt, att)
			}

		}
	}

	return json.Marshal(&struct {
		Alias
		CronSendTime string        `json:"send_time"`
		SendDate     string        `json:"send_date"`
		UpdateTime   string        `json:"update_time"`
		CreateTime   string        `json:"create_time"`
		Text         string        `json:"text"`
		Html         string        `json:"html"`
		Error        string        `json:"error"`
		Attachments  []attachments `json:"attachments"`
	}{
		Alias:        (Alias)(d),
		CronSendTime: d.CronSendTime.Format("2006-01-02 15:04:05"),
		UpdateTime:   d.UpdateTime.Format("2006-01-02 15:04:05"),
		CreateTime:   d.CreateTime.Format("2006-01-02 15:04:05"),
		SendDate:     d.SendDate.Format("2006-01-02 15:04:05"),
		Text:         d.Text.String,
		Html:         d.Html.String,
		Error:        d.Error.String,
		Attachments:  showAtt,
	})
}

func (d Email) ToTransObj() *parsemail.Email {

	return &parsemail.Email{
		From: &parsemail.User{
			Name:         d.FromName,
			EmailAddress: d.FromAddress,
		},
		To:          d.GetTos(),
		Subject:     d.Subject,
		Text:        []byte(d.Text.String),
		HTML:        []byte(d.Html.String),
		Sender:      d.GetSender(),
		ReplyTo:     d.GetReplyTo(),
		Bcc:         d.GetBcc(),
		Cc:          d.GetCc(),
		Attachments: d.GetAttachments(),
		Date:        d.SendDate.Format("2006-01-02 15:04:05"),
	}

}
