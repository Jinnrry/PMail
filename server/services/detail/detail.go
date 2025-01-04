package detail

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/Jinnrry/pmail/db"
	"github.com/Jinnrry/pmail/dto/parsemail"
	"github.com/Jinnrry/pmail/dto/response"
	"github.com/Jinnrry/pmail/models"
	"github.com/Jinnrry/pmail/services/list"
	"github.com/Jinnrry/pmail/utils/array"
	"github.com/Jinnrry/pmail/utils/context"
	"github.com/Jinnrry/pmail/utils/errors"
	log "github.com/sirupsen/logrus"
	"strings"
)
import . "xorm.io/builder"

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
	_, err = db.Instance.Select("*,1 as is_read").Table("email").Where("id=?", id).Get(&email)
	if err != nil {
		log.WithContext(ctx).Errorf("SQL error:%+v", err)
		return nil, err
	}

	email.IsRead = ue.IsRead

	if markRead && ue.IsRead == 0 {
		ue.IsRead = 1
		_, err = db.Instance.Where("id=?", ue.ID).Update(&ue)
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

func MakeRead(ctx *context.Context, emailId int, hadRead bool) {
	ue := models.UserEmail{
		UserID:  ctx.UserID,
		IsRead:  1,
		EmailID: emailId,
	}
	if !hadRead {
		ue.IsRead = 0
	}

	db.Instance.Where("email_id = ? and user_id=?", emailId, ctx.UserID).Cols("is_read").Update(&ue)
}

func FindUE(ctx *context.Context, groupName string, req list.ImapListReq, uid bool) []models.UserEmail {
	var ue []models.UserEmail
	if uid {
		err := db.Instance.Where(Eq{"id": req.UidList}).Find(&ue)
		if err != nil {
			log.WithContext(ctx).Errorf("SQL error:%+v", err)
		}
		return ue
	} else {
		sql := fmt.Sprintf("SELECT id,email_id, is_read from (SELECT id,email_id, is_read, ROW_NUMBER() OVER (ORDER BY id) AS serial_number FROM `user_email` WHERE (user_id = ? and status = ?)) a WHERE serial_number in (%s)",
			array.Join(req.UidList, ","),
		)
		switch groupName {
		case "INBOX":
			db.Instance.SQL(sql, ctx.UserID, 0).Find(&ue)
		case "Sent Messages":
			db.Instance.SQL(sql, ctx.UserID, 1).Find(&ue)
		case "Drafts":
			db.Instance.SQL(sql, ctx.UserID, 4).Find(&ue)
		case "Deleted Messages":
			db.Instance.SQL(sql, ctx.UserID, 3).Find(&ue)
		case "Junk":
			db.Instance.SQL(sql, ctx.UserID, 5).Find(&ue)
		default:
			groupNames := strings.Split(groupName, "/")
			groupName = groupNames[len(groupNames)-1]

			var group models.Group
			db.Instance.Table("group").Where("user_id=? and name=?", ctx.UserID, groupName).Get(&group)
			if group.ID == 0 {
				return nil
			}
			db.Instance.
				SQL(fmt.Sprintf(
					"SELECT * from (SELECT id,email_id, is_read, ROW_NUMBER() OVER (ORDER BY id) AS serial_number FROM `user_email` WHERE (user_id = ? and group_id = ?)) a WHERE serial_number in (%s)",
					array.Join(req.UidList, ","),
				)).
				Find(&ue, ctx.UserID, group.ID)
		}

		return ue

	}
}
