package detail

import (
	"database/sql"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"pmail/db"
	"pmail/dto/parsemail"
	"pmail/models"
	"pmail/utils/context"
	"strings"
)

func GetEmailDetail(ctx *context.Context, id int, markRead bool) (*models.Email, error) {
	// 获取邮件内容
	var email models.Email
	_, err := db.Instance.ID(id).Get(&email)
	if err != nil {
		log.WithContext(ctx).Errorf("SQL error:%+v", err)
		return nil, err
	}

	if markRead && email.IsRead == 0 {
		_, err = db.Instance.Exec(db.WithContext(ctx, "update email set is_read =1 where id =?"), email.Id)
		if err != nil {
			log.WithContext(ctx).Errorf("SQL error:%+v", err)
		}
	}

	// 将内容中的cid内容替换成url
	if email.Attachments != "" {
		var atts []parsemail.Attachment
		_ = json.Unmarshal([]byte(email.Attachments), &atts)
		for _, att := range atts {
			email.Html = sql.NullString{
				String: strings.ReplaceAll(email.Html.String, fmt.Sprintf("cid:%s", att.ContentID), fmt.Sprintf("/attachments/%d/%s", id, att.ContentID)),
			}
		}
	}

	return &email, nil
}
