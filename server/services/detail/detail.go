package detail

import (
	"database/sql"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"pmail/db"
	"pmail/dto/parsemail"
	"pmail/dto/response"
	"pmail/models"
	"pmail/utils/context"
	"pmail/utils/errors"
	"strings"
)

func GetEmailDetail(ctx *context.Context, id int, markRead bool) (*response.EmailResponseData, error) {
	// 先查是否是本人的邮件
	var ue models.UserEmail
	_, err := db.Instance.Where("email_id = ?", id).Get(&ue)
	if err != nil {
		log.Error(err)
	}
	if ue.ID == 0 && !ctx.IsAdmin {
		return nil, errors.New("Not authorized")
	}

	//获取邮件内容
	var email response.EmailResponseData
	_, err = db.Instance.ID(id).Get(&email)
	if err != nil {
		log.WithContext(ctx).Errorf("SQL error:%+v", err)
		return nil, err
	}

	email.IsRead = ue.IsRead

	if markRead && ue.IsRead == 0 {
		ue.IsRead = 1
		_, err = db.Instance.Update(&ue)
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
