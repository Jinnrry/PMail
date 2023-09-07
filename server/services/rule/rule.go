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
	if ctx == nil {
		err = db.Instance.Select(&res, "select * from rule order by sort desc")
	} else {
		err = db.Instance.Select(&res, db.WithContext(ctx, "select * from rule where user_id=? order by sort desc"), ctx.UserID)
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
	switch rule.Action {
	case dto.READ:
		email.IsRead = 1
	case dto.DELETE:
		email.Status = 3
	case dto.FORWARD:
		if strings.Contains(rule.Params, config.Instance.Domain) {
			log.WithContext(ctx).Errorf("Forward Error! loop forwarding!")
			return
		}

		err := send.Forward(nil, email, rule.Params)
		if err != nil {
			log.WithContext(ctx).Errorf("Forward Error:%v", err)
		}
	case dto.MOVE:
		email.GroupId = cast.ToInt(rule.Params)
	}
}
