package models

import (
	"database/sql"
	"encoding/json"
	"time"
)

type Email struct {
	Id           int            `xorm:"id pk unsigned int autoincr notnull" json:"id"`
	Type         int8           `xorm:"type tinyint(4) notnull default(0) comment('邮件类型，0:收到的邮件，1:发送的邮件')" json:"type"`
	Subject      string         `xorm:"subject varchar(1000) notnull default('') comment('邮件标题')" json:"subject"`
	ReplyTo      string         `xorm:"reply_to text comment('回复人')" json:"reply_to"`
	FromName     string         `xorm:"from_name varchar(50) notnull default('') comment('发件人名称')" json:"from_name"`
	FromAddress  string         `xorm:"from_address varchar(100) notnull default('') comment('发件人邮件地址')" json:"from_address"`
	To           string         `xorm:"to text comment('收件人地址')" json:"to"`
	Bcc          string         `xorm:"bcc text comment('密送')" json:"bcc"`
	Cc           string         `xorm:"cc text comment('抄送')" json:"cc"`
	Text         sql.NullString `xorm:"text text comment('文本内容')" json:"text"`
	Html         sql.NullString `xorm:"html mediumtext comment('html内容')" json:"html"`
	Sender       string         `xorm:"sender text comment('发送人')" json:"sender"`
	Attachments  string         `xorm:"attachments longtext comment('附件')" json:"attachments"`
	SPFCheck     int8           `xorm:"spf_check tinyint(1) comment('spf校验是否通过')" json:"spf_check"`
	DKIMCheck    int8           `xorm:"dkim_check tinyint(1) comment('dkim校验是否通过')" json:"dkim_check"`
	Status       int8           `xorm:"status tinyint(4) notnull default(0) comment('0未发送，1已发送，2发送失败')" json:"status"` // 0未发送，1已发送，2发送失败
	CronSendTime time.Time      `xorm:"cron_send_time comment('定时发送时间')" json:"cron_send_time"`
	UpdateTime   time.Time      `xorm:"update_time updated comment('更新时间')" json:"update_time"`
	SendUserID   int            `xorm:"send_user_id unsigned int  notnull default(0) comment('发件人用户id')" json:"send_user_id"`
	Size         int            `xorm:"size unsigned int  notnull default(1000) comment('邮件大小')" json:"size"`
	Error        sql.NullString `xorm:"error text comment('投递错误信息')" json:"error"`
	SendDate     time.Time      `xorm:"send_date comment('投递时间')" json:"send_date"`
	CreateTime   time.Time      `xorm:"create_time created" json:"create_time"`
}

func (d *Email) TableName() string {
	return "email"
}

type attachments struct {
	Filename    string
	ContentType string
	Index       int
	//Content     []byte
}

func (d *Email) MarshalJSON() ([]byte, error) {
	type Alias Email

	var allAtt = []attachments{}
	var showAtt = []attachments{}
	if d.Attachments != "" {
		_ = json.Unmarshal([]byte(d.Attachments), &allAtt)
		for i, att := range allAtt {
			att.Index = i
			//if att.ContentType == "application/octet-stream" {
			showAtt = append(showAtt, att)
			//}

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
		Alias:        (Alias)(*d),
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
