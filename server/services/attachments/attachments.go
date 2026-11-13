package attachments

import (
	"encoding/json"
	"mime"

	"github.com/Jinnrry/pmail/db"
	"github.com/Jinnrry/pmail/dto/parsemail"
	"github.com/Jinnrry/pmail/models"
	"github.com/Jinnrry/pmail/services/auth"
	"github.com/Jinnrry/pmail/utils/array"
	"github.com/Jinnrry/pmail/utils/context"
	log "github.com/sirupsen/logrus"
)

// InlineContentType reports whether raw is an image type safe to render inline.
func InlineContentType(raw string) (string, bool) {
	contentType, _, _ := mime.ParseMediaType(raw)
	if array.InArray(contentType, []string{
		"image/png", "image/jpeg", "image/gif",
		"image/webp", "image/bmp", "image/x-icon",
	}) {
		return contentType, true
	}
	return "application/octet-stream", false
}

func GetAttachments(ctx *context.Context, emailId int, cid string) *parsemail.Attachment {

	// 获取邮件内容
	var email models.Email
	_, err := db.Instance.ID(emailId).Get(&email)
	if err != nil {
		log.WithContext(ctx).Errorf("SQL error:%+v", err)
		return nil
	}

	// 检查权限
	if !auth.HasAuth(ctx, &email) {
		return nil
	}

	var atts []parsemail.Attachment
	_ = json.Unmarshal([]byte(email.Attachments), &atts)
	for i := range atts {
		if atts[i].ContentID == cid {
			return &atts[i]
		}
	}
	return nil
}

func GetAttachmentsByIndex(ctx *context.Context, emailId int, index int) (string, []byte) {

	// 获取邮件内容
	var email models.Email
	_, err := db.Instance.ID(emailId).Get(&email)
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
