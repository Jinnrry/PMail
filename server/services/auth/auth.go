package auth

import (
	log "github.com/sirupsen/logrus"
	"pmail/dto"
	"pmail/models"
	"pmail/mysql"
	"strings"
)

// HasAuth 检查当前用户是否有某个邮件的auth
func HasAuth(ctx *dto.Context, email *models.Email) bool {
	// 获取当前用户的auth
	var auth []models.UserAuth
	err := mysql.Instance.Select(&auth, mysql.WithContext(ctx, "select * from user_auth where user_id = ?"), ctx.UserInfo.ID)
	if err != nil {
		log.WithContext(ctx).Errorf("SQL error:%+v", err)
		return false
	}

	var hasAuth bool
	for _, userAuth := range auth {
		if userAuth.EmailAccount == "*" {
			hasAuth = true
			break
		} else if strings.Contains(email.Bcc, ctx.UserInfo.Account) || strings.Contains(email.Cc, ctx.UserInfo.Account) || strings.Contains(email.To, ctx.UserInfo.Account) {
			hasAuth = true
			break
		}
	}

	return hasAuth
}
