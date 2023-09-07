package match

import (
	log "github.com/sirupsen/logrus"
	"pmail/dto/parsemail"
	"pmail/utils/context"
	"regexp"
)

type RegexMatch struct {
	Rule  string
	Field string
}

func NewRegexMatch(field, rule string) *RegexMatch {
	return &RegexMatch{
		Rule:  rule,
		Field: field,
	}
}

func (r *RegexMatch) Match(ctx *context.Context, email *parsemail.Email) bool {
	content := getFieldContent(r.Field, email)

	match, err := regexp.MatchString(r.Rule, content)

	if err != nil {
		log.WithContext(ctx).Errorf("rule regex error %v", err)
	}

	return match
}
