package del_email

import (
	log "github.com/sirupsen/logrus"
	"pmail/consts"
	"pmail/db"
	"pmail/models"
	"pmail/utils/context"
)
import . "xorm.io/builder"

func DelEmail(ctx *context.Context, ids []int) error {

	if len(ids) == 0 {
		return nil
	}

	where, params, err := ToSQL(Eq{"user_id": ctx.UserID}.And(Eq{"email_id": ids}))

	if err != nil {
		log.Errorf("del email err: %v", err)
		return err
	}

	_, err = db.Instance.Table(&models.UserEmail{}).Where(where, params...).Update(map[string]interface{}{"status": consts.EmailStatusDel})
	if err != nil {
		log.Errorf("del email err: %v", err)
	}
	return err
}

func DelEmailI64(ctx *context.Context, ids []int64) error {

	if len(ids) == 0 {
		return nil
	}

	where, params, err := ToSQL(Eq{"user_id": ctx.UserID}.And(Eq{"email_id": ids}))

	if err != nil {
		log.Errorf("del email err: %v", err)
		return err
	}

	_, err = db.Instance.Table(&models.UserEmail{}).Where(where, params...).Update(map[string]interface{}{"status": consts.EmailStatusDel})
	if err != nil {
		log.Errorf("del email err: %v", err)
	}
	return err
}
