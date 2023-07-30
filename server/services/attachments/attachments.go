package attachments

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"pmail/dto"
	"pmail/dto/parsemail"
	"pmail/models"
	"pmail/mysql"
	"pmail/services/auth"
)

func GetAttachments(ctx *dto.Context, emailId int, cid string) (string, []byte) {

	// 获取邮件内容
	var email models.Email
	err := mysql.Instance.Get(&email, mysql.WithContext(ctx, "select * from email where id = ?"), emailId)
	if err != nil {
		log.WithContext(ctx).Errorf("SQL error:%+v", err)
		return "", nil
	}

	// 检查权限
	if !auth.HasAuth(ctx, &email) {
		return "", nil
	}

	var atts []parsemail.Attachment
	_ = json.Unmarshal([]byte(email.Attachments), &atts)
	for _, att := range atts {
		if att.ContentID == cid {
			return att.ContentType, att.Content
		}
	}
	return "", nil
}

func GetAttachmentsByIndex(ctx *dto.Context, emailId int, index int) (string, []byte) {

	// 获取邮件内容
	var email models.Email
	err := mysql.Instance.Get(&email, mysql.WithContext(ctx, "select * from email where id = ?"), emailId)
	if err != nil {
		log.WithContext(ctx).Errorf("SQL error:%+v", err)
		return "", nil
	}

	// 检查权限
	if !auth.HasAuth(ctx, &email) {
		return "", nil
	}

	var atts []parsemail.Attachment
	_ = json.Unmarshal([]byte(email.Attachments), &atts)

	if len(atts) > index {
		return atts[index].Filename, atts[index].Content
	}
	return "", nil
}
