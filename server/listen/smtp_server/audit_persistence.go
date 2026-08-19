package smtp_server

import (
	"database/sql"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"strings"

	"github.com/Jinnrry/pmail/config"
	"github.com/Jinnrry/pmail/consts"
	"github.com/Jinnrry/pmail/db"
	"github.com/Jinnrry/pmail/dto/parsemail"
	"github.com/Jinnrry/pmail/hooks"
	"github.com/Jinnrry/pmail/hooks/framework"
	"github.com/Jinnrry/pmail/models"
	"github.com/Jinnrry/pmail/utils/array"
	"github.com/Jinnrry/pmail/utils/async"
	"github.com/Jinnrry/pmail/utils/context"
	smtp "github.com/emersion/go-smtp"
	log "github.com/sirupsen/logrus"
	"xorm.io/builder"
	"xorm.io/xorm"
)

const deliveryPendingError = "delivery result pending; outcome unknown after interruption"

type outgoingSender func(*context.Context, *parsemail.Email) (error, map[string]error)

func (s *Session) deliverOutgoingEmail(email *parsemail.Email, sender outgoingSender) error {
	ctx := s.Ctx
	_, saved, err := saveEmail(ctx, email.Size, email, ctx.UserID, int(consts.EmailTypeSend), nil, true, true)
	if err != nil {
		log.WithContext(ctx).Errorf("CRITICAL audit_event=initial_persist_failed direction=outbound msg_id=%q error=%v", email.MsgID, err)
		return localPersistenceSMTPError()
	}

	deliveryErr, domainErrors := sender(ctx, email)
	runSendAfterHooks(ctx, email, domainErrors)

	targetStatus := consts.EmailStatusSent
	var errorText any
	if deliveryErr != nil {
		targetStatus = consts.EmailStatusFail
		errorText = deliveryErr.Error()
	}

	if err = persistDeliveryResult(ctx, saved.Id, targetStatus, errorText); err != nil {
		log.WithContext(ctx).Errorf(
			"CRITICAL audit_event=delivery_result_persist_failed email_id=%d msg_id=%q target_status=%d downstream_accepted=%t error=%v",
			saved.Id, saved.MsgID, targetStatus, deliveryErr == nil, err,
		)
	}

	// If the remote MX accepted DATA, returning a local 451 here would cause a
	// duplicate delivery. The persistence failure is retained as pending state
	// and a CRITICAL log; the SMTP result must still reflect the remote result.
	return downstreamDeliverySMTPError(deliveryErr, domainErrors)
}

func runSendAfterHooks(ctx *context.Context, email *parsemail.Email, domainErrors map[string]error) {
	process := async.New(ctx)
	for _, hook := range hooks.HookList {
		if hook == nil {
			continue
		}
		process.WaitProcess(func(value any) {
			value.(framework.EmailHook).SendAfter(ctx, email, domainErrors)
		}, hook)
	}
	process.Wait()
}

func persistDeliveryResult(ctx *context.Context, emailID int, status int8, errorText any) error {
	return db.Transaction(db.Instance, func(session *xorm.Session) error {
		result, err := session.Exec(db.WithContext(ctx, `update email
			set status=?, error=?, update_time=CURRENT_TIMESTAMP
			where id=?`), status, errorText, emailID)
		if err != nil {
			return err
		}
		if err = requireOneAffected(result, "update email delivery result"); err != nil {
			return err
		}

		result, err = session.Exec(db.WithContext(ctx, `update user_email
			set status=? where email_id=?`), status, emailID)
		if err != nil {
			return err
		}
		return requireOneAffected(result, "update user_email delivery result")
	})
}

func requireOneAffected(result sql.Result, operation string) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: read affected rows: %w", operation, err)
	}
	if affected != 1 {
		return fmt.Errorf("%s: affected %d rows, want 1", operation, affected)
	}
	return nil
}

func localPersistenceSMTPError() error {
	return &smtp.SMTPError{
		Code:         451,
		EnhancedCode: smtp.EnhancedCode{4, 3, 0},
		Message:      "Temporary local persistence failure",
	}
}

func persistInitialAudit(ctx *context.Context, modelEmail models.Email, userIDs []int) (*models.Email, error) {
	var saved models.Email
	err := db.Transaction(db.Instance, func(session *xorm.Session) error {
		candidate := modelEmail
		candidate.Id = 0
		if _, err := session.Insert(&candidate); err != nil {
			return err
		}
		if candidate.Id <= 0 {
			return stderrors.New("email insert returned no id")
		}
		for _, userID := range userIDs {
			relation := models.UserEmail{EmailID: candidate.Id, UserID: userID, Status: candidate.Status}
			if _, err := session.Insert(&relation); err != nil {
				return err
			}
		}
		saved = candidate
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &saved, nil
}

func resolveIncomingUsers(ctx *context.Context, email *parsemail.Email, emailType int, reallyTo []string, SPFStatus, dkimStatus bool) ([]*models.User, bool, error) {
	if emailType != int(consts.EmailTypeReceive) {
		return nil, false, nil
	}

	accounts := incomingAccounts(email, reallyTo)
	var users []*models.User
	if len(accounts) > 0 {
		where, params, err := builder.ToSQL(builder.In("LOWER(account)", accounts))
		if err != nil {
			return nil, false, err
		}
		if err = db.Instance.Table(&models.User{}).Where(where, params...).Find(&users); err != nil {
			return nil, false, err
		}
	}
	if len(users) > 0 {
		return users, false, nil
	}

	if DropUnknownRecipientEmails &&
		((config.Instance.SpamFilterLevel == 1 && !SPFStatus && !dkimStatus) ||
			(config.Instance.SpamFilterLevel == 2 && !SPFStatus) ||
			(config.Instance.SpamFilterLevel == 3 && !dkimStatus)) {
		log.WithContext(ctx).Infoln("垃圾邮件，拒信")
		log.WithContext(ctx).Infof("收件人不存在且DKIM验证失败，丢弃邮件: %s -> %v", email.From.EmailAddress, accounts)
		return nil, true, nil
	}

	log.WithContext(ctx).Infof("收件人不存在但DKIM验证通过，转交管理员: %s -> %v", email.From.EmailAddress, accounts)
	if err := db.Instance.Table(&models.User{}).Where("is_admin=1").Find(&users); err != nil {
		return nil, false, err
	}
	if len(users) == 0 {
		return nil, false, stderrors.New("no administrator available for unknown recipient")
	}
	return users, false, nil
}

func incomingAccounts(email *parsemail.Email, reallyTo []string) []string {
	var accounts []string
	if len(reallyTo) > 0 {
		for _, recipient := range reallyTo {
			account := parsemail.BuilderUser(recipient)
			if account == nil {
				continue
			}
			name, domain := account.GetDomainAccount()
			if array.InArray(domain, config.Instance.Domains) && name != "" {
				accounts = append(accounts, strings.ToLower(name))
			}
		}
		return accounts
	}
	for _, recipients := range [][]*parsemail.User{email.To, email.Cc, email.Bcc} {
		for _, user := range recipients {
			account, _ := user.GetDomainAccount()
			if account != "" {
				accounts = append(accounts, strings.ToLower(account))
			}
		}
	}
	return accounts
}

func boolInt8(value bool) int8 {
	if value {
		return 1
	}
	return 0
}

func json2string(value any) string {
	data, _ := json.Marshal(value)
	return string(data)
}
