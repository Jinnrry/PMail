package del_email

import (
	"fmt"
	"pmail/db"
	"pmail/dto"
	"pmail/models"
	"pmail/services/auth"
	"pmail/utils/array"
	"pmail/utils/errors"
)

func DelEmail(ctx *dto.Context, ids []int) error {
	var emails []*models.Email

	db.Instance.Select(&emails, db.WithContext(ctx, fmt.Sprintf("select * from email where id in (%s)", array.Join(ids, ","))))

	for _, email := range emails {
		// 检查是否有权限
		hasAuth := auth.HasAuth(ctx, email)
		if !hasAuth {
			return errors.New("No Auth!")
		}
	}

	_, err := db.Instance.Exec(db.WithContext(ctx, fmt.Sprintf("delete from email where id in (%s)", array.Join(ids, ","))))
	if err != nil {
		return errors.Wrap(err)
	}

	return nil
}
