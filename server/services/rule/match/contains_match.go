package match

import (
	"pmail/dto/parsemail"
	"pmail/utils/context"
	"strings"
)

type ContainsMatch struct {
	Rule  string
	Field string
}

func NewContainsMatch(field, rule string) *ContainsMatch {
	return &ContainsMatch{
		Rule:  rule,
		Field: field,
	}
}

func (r *ContainsMatch) Match(ctx *context.Context, email *parsemail.Email) bool {
	content := getFieldContent(r.Field, email)
	return strings.Contains(content, r.Rule)
}
