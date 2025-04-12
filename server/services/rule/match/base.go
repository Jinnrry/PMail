package match

import (
	"github.com/Jinnrry/pmail/dto/parsemail"
	"github.com/Jinnrry/pmail/utils/context"
)

const (
	RuleTypeRegex    = "regex"
	RuleTypeContains = "contains"
	RuleTypeEq       = "equal"
)

type Match interface {
	Match(ctx *context.Context, email *parsemail.Email) bool
}

func buildUsers(users []*parsemail.User) string {
	ret := ""
	for i, u := range users {
		if i != 0 {
			ret += ","
		}
		ret += u.EmailAddress
	}
	return ret
}

func getFieldContent(field string, email *parsemail.Email) string {
	switch field {
	case "ReplyTo":
		return buildUsers(email.ReplyTo)
	case "From":
		return email.From.EmailAddress
	case "Subject":
		return email.Subject
	case "To":
		return buildUsers(email.To)
	case "Bcc":
		return buildUsers(email.Bcc)
	case "Cc":
		return buildUsers(email.Cc)
	case "Text":
		return string(email.Text)
	case "Html":
		return string(email.HTML)
	case "Sender":
		return email.Sender.EmailAddress
	case "Content":
		b := string(email.HTML)
		b2 := string(email.Text)
		return b + b2
	}
	return ""
}
