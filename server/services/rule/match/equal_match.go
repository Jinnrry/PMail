package match

import (
	"github.com/Jinnrry/pmail/dto/parsemail"
	"github.com/Jinnrry/pmail/utils/context"
)

type EqualMatch struct {
	Rule  string
	Field string
}

func NewEqualMatch(field, rule string) *EqualMatch {
	return &EqualMatch{
		Rule:  rule,
		Field: field,
	}
}

func (r *EqualMatch) Match(ctx *context.Context, email *parsemail.Email) bool {
	content := getFieldContent(r.Field, email)
	return content == r.Rule
}
