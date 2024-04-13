package rule

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"pmail/config"
	"pmail/db"
	"pmail/dto"
	"pmail/dto/parsemail"
	"pmail/models"
	"pmail/services/rule/match"
	"pmail/utils/context"
	"pmail/utils/send"
	"strings"
)

func GetAllRules(ctx *context.Context) []*dto.Rule {
	var res []*models.Rule
	var err error
	if ctx == nil || ctx.UserID == 0 {
		err = db.Instance.Decr("sort").Find(&res)
	} else {
		err = db.Instance.Where("user_id=?", ctx.UserID).Decr("sort").Find(&res)
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

func DoRule(ctx *context.Context, rule *dto.Rule, email *parsemail.Email) {
	log.WithContext(ctx).Debugf("执行规则:%s", rule.Name)

	switch rule.Action {
	case dto.READ:
		email.IsRead = 1
		if email.MessageId > 0 {
			db.Instance.Exec(db.WithContext(ctx, "update email set is_read=1 where id =?"), email.MessageId)
		}
	case dto.DELETE:
		email.Status = 3
		if email.MessageId > 0 {
			db.Instance.Exec(db.WithContext(ctx, "update email set status=3 where id =?"), email.MessageId)
		}
	case dto.FORWARD:
		if strings.Contains(rule.Params, config.Instance.Domain) {
			log.WithContext(ctx).Errorf("Forward Error! loop forwarding!")
			return
		}
		err := send.Forward(ctx, email, rule.Params)
		if err != nil {
			log.WithContext(ctx).Errorf("Forward Error:%v", err)
		}
	case dto.MOVE:
		email.GroupId = cast.ToInt(rule.Params)
		if email.MessageId > 0 {
			db.Instance.Exec(db.WithContext(ctx, "update email set group_id=? where id =?"), email.GroupId, email.MessageId)
		}
	}

}
