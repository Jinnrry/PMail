package rule

import (
	"fmt"
	"strings"

	"github.com/Jinnrry/pmail/config"
	"github.com/Jinnrry/pmail/consts"
	"github.com/Jinnrry/pmail/db"
	"github.com/Jinnrry/pmail/dto"
	"github.com/Jinnrry/pmail/dto/parsemail"
	"github.com/Jinnrry/pmail/models"
	"github.com/Jinnrry/pmail/services/rule/match"
	"github.com/Jinnrry/pmail/utils/context"
	"github.com/Jinnrry/pmail/utils/send"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
)

func GetAllRules(ctx *context.Context, userId int) []*dto.Rule {
	var res []*models.Rule
	var err error
	if userId == 0 {
		return nil
	} else {
		err = db.Instance.Where("user_id=?", userId).Decr("sort").Find(&res)
	}

	if err != nil {
		log.WithContext(ctx).Errorf("sqlERror :%v", err)
	}
	var ret []*dto.Rule
	for _, rule := range res {
		ret = append(ret, (&dto.Rule{}).Decode(rule))
	}

	return ret
}

func MatchRule(ctx *context.Context, rule *dto.Rule, email *parsemail.Email) bool {

	for _, r := range rule.Rules {
		var m match.Match

		switch r.Type {
		case match.RuleTypeRegex:
			m = match.NewRegexMatch(r.Field, r.Rule)
		case match.RuleTypeContains:
			m = match.NewContainsMatch(r.Field, r.Rule)
		case match.RuleTypeEq:
			m = match.NewEqualMatch(r.Field, r.Rule)
		}
		if m == nil {
			continue
		}

		if !m.Match(ctx, email) {
			return false
		}
	}

	return true
}

func DoRule(ctx *context.Context, rule *dto.Rule, email *parsemail.Email, user *models.User, rawEmailData []byte) {
	log.WithContext(ctx).Debugf("执行规则:%s", rule.Name)

	switch rule.Action {
	case dto.READ:
		if email.MessageId > 0 {
			_, err := db.Instance.Table(&models.UserEmail{}).Where("email_id=? and user_id=?", email.MessageId, rule.UserId).Cols("is_read").Update(map[string]interface{}{"is_read": 1})
			if err != nil {
				log.WithContext(ctx).Errorf("sqlERror :%v", err)
			}
		}
	case dto.DELETE:
		_, err := db.Instance.Table(&models.UserEmail{}).Where("email_id=? and user_id=?", email.MessageId, rule.UserId).Cols("status").Update(map[string]interface{}{"status": consts.EmailStatusDel})
		if err != nil {
			log.WithContext(ctx).Errorf("sqlERror :%v", err)
		}
	case dto.FORWARD:
		err := doForward(ctx, email, rule.Params, user, rawEmailData)
		if err != nil {
			log.WithContext(ctx).Errorf("Forward Error:%v", err)
		} else {
			log.WithContext(ctx).Infof("Forward Success:%s@%s -> %s", user.Account, config.Instance.Domains[0], rule.Params)
		}
	case dto.MOVE:
		doMove(ctx, rule, email, user)
	}

}

func doForward(ctx *context.Context, email *parsemail.Email, forwardAddress string, user *models.User, rawEmailData []byte) error {
	forwardUser := parsemail.BuilderUser(forwardAddress)
	if forwardUser == nil || forwardUser.EmailAddress == "" {
		return fmt.Errorf("invalid forward address: %s", forwardAddress)
	}

	account, domain := forwardUser.GetDomainAccount()
	if account == "" || domain == "" {
		return fmt.Errorf("invalid forward address: %s", forwardAddress)
	}

	if isLocalDomain(domain) {
		if strings.EqualFold(account, user.Account) {
			return fmt.Errorf("loop forwarding to self: %s", forwardAddress)
		}
		return forwardToLocalUser(ctx, email, account, forwardAddress)
	}

	if len(rawEmailData) > 0 {
		err := send.ForwardRaw(ctx, email, rawEmailData, forwardAddress, user)
		if err == nil {
			return nil
		}
		log.WithContext(ctx).Warnf("Raw forward failed, retry remail forward:%v", err)
	}
	return send.Forward(ctx, email, forwardAddress, user)
}

func forwardToLocalUser(ctx *context.Context, email *parsemail.Email, account, forwardAddress string) error {
	if email.MessageId <= 0 {
		return fmt.Errorf("email has not been saved before local forwarding")
	}

	var user models.User
	has, err := db.Instance.Table(&models.User{}).
		Where("LOWER(account)=LOWER(?) and disabled=0", account).
		Get(&user)
	if err != nil {
		return err
	}
	if !has {
		return fmt.Errorf("local forward user not found: %s", forwardAddress)
	}

	var ue models.UserEmail
	has, err = db.Instance.Table(&models.UserEmail{}).
		Where("email_id=? and user_id=?", email.MessageId, user.ID).
		Get(&ue)
	if err != nil {
		return err
	}
	if has {
		return nil
	}

	_, err = db.Instance.Insert(&models.UserEmail{
		UserID:  user.ID,
		EmailID: cast.ToInt(email.MessageId),
		Status:  cast.ToInt8(email.Status),
	})
	return err
}

func isLocalDomain(domain string) bool {
	for _, localDomain := range config.Instance.Domains {
		if strings.EqualFold(domain, localDomain) {
			return true
		}
	}
	return strings.EqualFold(domain, config.Instance.Domain)
}

func doMove(ctx *context.Context, rule *dto.Rule, email *parsemail.Email, user *models.User) {

	groupId := cast.ToInt(rule.Params)
	switch groupId {
	case models.INBOX:
		_, err := db.Instance.Table(&models.Email{}).Where("id=?", email.MessageId).
			Cols("type").Update(map[string]interface{}{"type": consts.EmailTypeReceive})
		if err != nil {
			log.WithContext(ctx).Errorf("sqlERror :%v", err)
		}

		_, err = db.Instance.Table(&models.UserEmail{}).Where("email_id=? and user_id=?", email.MessageId, rule.UserId).
			Cols("group_id", "status").Update(map[string]interface{}{"group_id": 0, "status": 0})
		if err != nil {
			log.WithContext(ctx).Errorf("sqlERror :%v", err)
		}
	case models.Sent:
		_, err := db.Instance.Table(&models.Email{}).Where("id=?", email.MessageId).
			Cols("type").Update(map[string]interface{}{"type": consts.EmailTypeSend})
		if err != nil {
			log.WithContext(ctx).Errorf("sqlERror :%v", err)
		}

		_, err = db.Instance.Table(&models.UserEmail{}).Where("email_id=? and user_id=?", email.MessageId, rule.UserId).
			Cols("group_id", "status").Update(map[string]interface{}{"group_id": 0, "status": 0})
		if err != nil {
			log.WithContext(ctx).Errorf("sqlERror :%v", err)
		}

	case models.Drafts:
		_, err := db.Instance.Table(&models.UserEmail{}).Where("email_id=? and user_id=?", email.MessageId, rule.UserId).
			Cols("group_id", "status").Update(map[string]interface{}{"group_id": 0, "status": consts.EmailStatusDrafts})
		if err != nil {
			log.WithContext(ctx).Errorf("sqlERror :%v", err)
		}

	case models.Deleted:
		_, err := db.Instance.Table(&models.UserEmail{}).Where("email_id=? and user_id=?", email.MessageId, rule.UserId).
			Cols("group_id", "status").Update(map[string]interface{}{"group_id": 0, "status": consts.EmailStatusDel})
		if err != nil {
			log.WithContext(ctx).Errorf("sqlERror :%v", err)
		}
	case models.Junk:
		_, err := db.Instance.Table(&models.UserEmail{}).Where("email_id=? and user_id=?", email.MessageId, rule.UserId).
			Cols("group_id", "status").Update(map[string]interface{}{"group_id": 0, "status": consts.EmailStatusJunk})
		if err != nil {
			log.WithContext(ctx).Errorf("sqlERror :%v", err)
		}

	default:

		_, err := db.Instance.Table(&models.UserEmail{}).Where("email_id=? and user_id=?", email.MessageId, rule.UserId).Cols("group_id").Update(map[string]interface{}{"group_id": groupId})
		if err != nil {
			log.WithContext(ctx).Errorf("sqlERror :%v", err)
		}
	}

}
