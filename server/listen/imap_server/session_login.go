package imap_server

import (
	"database/sql"
	"github.com/Jinnrry/pmail/db"
	"github.com/Jinnrry/pmail/models"
	"github.com/Jinnrry/pmail/utils/errors"
	"github.com/Jinnrry/pmail/utils/password"
	"github.com/emersion/go-imap/v2"
	log "github.com/sirupsen/logrus"
	"strings"
)

func (s *serverSession) Login(username, pwd string) error {
	if strings.Contains(username, "@") {
		args := strings.Split(username, "@")
		username = args[0]
	}

	var user models.User

	encodePwd := password.Encode(pwd)

	_, err := db.Instance.Where("account =? and password =? and disabled = 0", username, encodePwd).Get(&user)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.WithContext(s.ctx).Errorf("%+v", err)
	}

	if user.ID > 0 {
		s.status = AUTHORIZED

		s.ctx.UserID = user.ID
		s.ctx.UserName = user.Name
		s.ctx.UserAccount = user.Account
		log.WithContext(s.ctx).Debug("Login successful")

		return nil
	}

	log.WithContext(s.ctx).Info("user not found")
	return &imap.Error{
		Type: imap.StatusResponseTypeNo,
		Code: imap.ResponseCodeAuthenticationFailed,
		Text: "Invalid credentials (Failure)",
	}
}
