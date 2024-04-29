package del_email

import (
	"pmail/db"
	"pmail/models"
	"pmail/services/auth"
	"pmail/utils/context"
	"pmail/utils/errors"
	"xorm.io/builder"
)

func DelEmail(ctx *context.Context, ids []int) error {
	var emails []*models.Email

	err := db.Instance.Table("email").Where(builder.In("id", ids)).Find(&emails)
	if err != nil {
		return errors.Wrap(err)
	}
	for _, email := range emails {
		// 检查是否有权限
		hasAuth := auth.HasAuth(ctx, email)
		if !hasAuth {
			return errors.New("No Auth!")
		}
		email.Status = 3
	}

	_, err = db.Instance.Table("email").Where(builder.In("id", ids)).Cols("status").Update(map[string]interface{}{"status": 3})

	if err != nil {
		return errors.Wrap(err)
	}

	return nil
}
