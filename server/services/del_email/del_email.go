package del_email

import (
	"github.com/Jinnrry/pmail/consts"
	"github.com/Jinnrry/pmail/db"
	"github.com/Jinnrry/pmail/models"
	"github.com/Jinnrry/pmail/utils/context"
	"github.com/spf13/cast"
	"xorm.io/xorm"
)

func DelEmail(ctx *context.Context, ids []int64, forcedDel bool) error {
	session := db.Instance.NewSession()
	defer session.Close()
	if err := session.Begin(); err != nil {
		return err
	}
	for _, id := range ids {
		err := deleteOne(ctx, session, cast.ToInt64(id), forcedDel)
		if err != nil {
			session.Rollback()
			return err
		}
	}
	return session.Commit()
}

type num struct {
	Num int `xorm:"num"`
}

func deleteOne(ctx *context.Context, session *xorm.Session, id int64, forcedDel bool) error {
	if !forcedDel {
		_, err := session.Table(&models.UserEmail{}).Where("email_id=? and user_id=?", id, ctx.UserID).Update(map[string]interface{}{"status": consts.EmailStatusDel})
		return err
	}
	// 先删除关联关系
	var ue models.UserEmail
	_, err := session.Table(&models.UserEmail{}).Where("email_id=? and user_id=?", id, ctx.UserID).Delete(&ue)
	if err != nil {
		return err
	}
	// 检查email是否还有人有权限
	var Num num
	_, err = session.Table(&models.UserEmail{}).Select("count(1) as num").Where("email_id=? ", id).Get(&Num)
	if err != nil {
		return err
	}
	if Num.Num == 0 {
		var email models.Email
		_, err = session.Table(&email).Where("id=?", id).Delete(&email)

	}
	return err
}
