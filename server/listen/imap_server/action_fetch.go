package imap_server

import (
	"fmt"
	"github.com/Jinnrry/pmail/dto/parsemail"
	"github.com/Jinnrry/pmail/dto/response"
	"github.com/Jinnrry/pmail/services/detail"
	"github.com/Jinnrry/pmail/services/list"
	"github.com/Jinnrry/pmail/utils/context"
	"github.com/Jinnrry/pmail/utils/goimap"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"strings"
)

func (a action) Fetch(session *goimap.Session, mailIds, commands string, uid bool) goimap.CommandResponse {
	log.Infof("%s,%s,%s", "Fetch", mailIds, commands)
	if session.CurrentPath == "" {
		return goimap.CommandResponse{
			Type:    goimap.BAD,
			Message: "Please Select Mailbox!",
		}
	}

	offset := 0
	limit := 0

	if strings.Contains(mailIds, ":") {
		args := strings.Split(mailIds, ":")
		offset = cast.ToInt(args[0])
		limit = cast.ToInt(args[1])
	} else {
		offset = cast.ToInt(mailIds)
		limit = 1
	}
	if offset > 0 {
		offset -= 1
	}
	emailList := list.GetEmailListByGroup(session.Ctx.(*context.Context), session.CurrentPath, offset, limit)
	ret := goimap.CommandResponse{}

	commandArg := splitCommand(commands, uid)

	for i, email := range emailList {
		buildResponse(session.Ctx.(*context.Context), offset+i+1, email, commandArg, &ret)
	}

	ret.Message = "FETCH Completed"

	return ret
}

func buildResponse(ctx *context.Context, no int, email *response.EmailResponseData, commands []string, ret *goimap.CommandResponse) {
	retStr := ""
	for _, command := range commands {
		switch command {
		case "INTERNALDATE":
			if retStr != "" {
				retStr += " "
			}
			retStr += fmt.Sprintf(`INTERNALDATE "%s"`, email.CreateTime.Format("2-Jan-2006 15:04:05 -0700"))
		case "UID":
			if retStr != "" {
				retStr += " "
			}
			retStr += fmt.Sprintf(`UID %d`, no)
		case "RFC822.SIZE":
			if retStr != "" {
				retStr += " "
			}
			retStr += fmt.Sprintf(`RFC822.SIZE %d`, email.Size)
		case "FLAGS":
			if retStr != "" {
				retStr += " "
			}
			if email.IsRead == 1 {
				retStr += `FLAGS (\Seen)`
			} else {
				retStr += `FLAGS ()`
			}
		default:
			if strings.HasPrefix(command, "BODY") {
				if retStr != "" {
					retStr += " "
				}

				retStr += strings.Replace(command, ".PEEK", "", 1) + buildBody(ctx, command, email)
			}
		}
	}
	ret.Data = append(ret.Data, fmt.Sprintf("* %d FETCH (%s)", no, retStr))
}

type item struct {
	content string
	name    string
}

func buildBody(ctx *context.Context, command string, email *response.EmailResponseData) string {
	if !strings.HasPrefix(command, "BODY.PEEK") && email.IsRead == 0 {
		detail.MakeRead(ctx, email.Id)
	}
	ret := ""
	fields := []string{}
	if strings.Contains(command, "HEADER.FIELDS") {
		args := strings.Split(command, "(")
		data := strings.Split(args[1], ")")
		fields = strings.Split(data[0], " ")
	}
	emailContent := parsemail.NewEmailFromModel(email.Email).BuildBytes(ctx, false)
	headerMap := map[string]*item{}
	var key string
	var isContent bool
	content := ""

	for _, line := range strings.Split(string(emailContent), "\r\n") {
		if line == "" {
			isContent = true
		}
		if isContent {
			content += line + "\r\n"
		} else {
			if !strings.HasPrefix(line, " ") {
				args := strings.SplitN(line, ":", 2)
				key = strings.ToTitle(args[0])
				headerMap[key] = &item{strings.TrimSpace(args[1]), args[0]}
			} else {
				headerMap[key].content += fmt.Sprintf("\r\n%s", line)
			}
		}

	}

	if len(fields) == 0 {
		for _, v := range headerMap {
			ret += fmt.Sprintf("%s: %s\r\n", v.name, v.content)
		}
		ret += content
	} else {
		for _, field := range fields {
			field = strings.Trim(field, `" `)

			key := strings.ToTitle(field)

			if headerMap[key] != nil {
				ret += fmt.Sprintf("%s: %s\r\n", headerMap[key].name, headerMap[key].content)
			}
		}
	}

	size := len([]byte(ret)) + 2

	return fmt.Sprintf(" {%d}\r\n%s\r\n", size, ret)
}

func splitCommand(commands string, uid bool) []string {
	var ret []string
	if uid {
		ret = append(ret, "UID")
	}

	commands = strings.Trim(commands, "() ")

	for i := 0; i < 1000; i++ {
		if commands == "" {
			break
		}
		if !strings.HasPrefix(commands, "BODY") {
			args := strings.SplitN(commands, " ", 2)
			if len(args) >= 2 {
				commands = strings.TrimSpace(args[1])
			} else {
				commands = ""
			}
			ret = append(ret, args[0])
		} else {
			item := ""
			if strings.HasPrefix(commands, "BODY.PEEK") {
				commands = strings.TrimPrefix(commands, "BODY.PEEK")
				item += "BODY.PEEK"
			} else if strings.HasPrefix(commands, "BODY") {
				commands = strings.TrimPrefix(commands, "BODY")
				item += "BODY"
			}
			if commands[0] == '[' {
				args := strings.SplitN(commands, "]", 2)
				item += args[0] + "]"
				ret = append(ret, item)
				commands = strings.TrimSpace(args[1])
			}
		}
	}

	return ret
}
