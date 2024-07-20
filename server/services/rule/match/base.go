package match

import (
	"encoding/json"
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

func getFieldContent(field string, email *parsemail.Email) string {
	switch field {
	case "ReplyTo":
		b, _ := json.Marshal(email.ReplyTo)
		return string(b)
	case "From":
		b, _ := json.Marshal(email.From)
		return string(b)
	case "Subject":
		return email.Subject
	case "To":
		b, _ := json.Marshal(email.To)
		return string(b)
	case "Bcc":
		b, _ := json.Marshal(email.Bcc)
		return string(b)
	case "Cc":
		b, _ := json.Marshal(email.Cc)
		return string(b)
	case "Text":
		return string(email.Text)
	case "Html":
		return string(email.HTML)
	case "Sender":
		b, _ := json.Marshal(email.Sender)
		return string(b)
	case "Content":
		b := string(email.HTML)
		b2 := string(email.Text)
		return b + b2
	}
	return ""
}
