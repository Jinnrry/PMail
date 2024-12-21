package imap_server

import (
	"database/sql"
	"fmt"
	"github.com/Jinnrry/pmail/db"
	"github.com/Jinnrry/pmail/models"
	"github.com/Jinnrry/pmail/services/group"
	"github.com/Jinnrry/pmail/utils/array"
	"github.com/Jinnrry/pmail/utils/context"
	"github.com/Jinnrry/pmail/utils/errors"
	"github.com/Jinnrry/pmail/utils/goimap"
	"github.com/Jinnrry/pmail/utils/id"
	"github.com/Jinnrry/pmail/utils/password"
	log "github.com/sirupsen/logrus"
	"strings"
	"sync"
)

var idlePool sync.Map

func init() {
	idlePool = sync.Map{}
}

// PushMsgByIDLE 向IMAP客户端通知新邮件消息
func PushMsgByIDLE(ctx *context.Context, account string, unionId string) error {
	sess, ok := idlePool.Load(account)
	if ok {
		sSessions, ok2 := sess.([]*goimap.Session)
		if !ok2 {
			idlePool.Delete(account)
			return nil
		}
		newPool := []*goimap.Session{}
		for _, sSession := range sSessions {
			if sSession.IN_IDLE && sSession.Conn != nil {
				fmt.Fprintf(sSession.Conn, fmt.Sprintf("* %s EXISTS", unionId))
				newPool = append(newPool, sSession)
			}
		}
		if len(newPool) == 0 {
			idlePool.Delete(account)
		} else {
			idlePool.Store(account, newPool)
		}
	}
	return nil
}

type action struct{}

func (a action) Create(session *goimap.Session, path string) goimap.CommandResponse {
	log.Infof("%s,%s", "Create", path)
	return goimap.CommandResponse{}
}

func (a action) Delete(session *goimap.Session, path string) goimap.CommandResponse {
	log.Infof("%s,%s", "Delete", path)
	return goimap.CommandResponse{}
}

func (a action) Rename(session *goimap.Session, oldPath, newPath string) goimap.CommandResponse {
	log.Infof("%s,%s,%s", "Rename", oldPath, newPath)
	return goimap.CommandResponse{}
}

func (a action) List(session *goimap.Session, basePath, template string) goimap.CommandResponse {
	log.Infof("%s,%s,%s", "List", basePath, template)
	var ret []string
	if basePath == "" && template == "" {
		ret = append(ret, `* LIST (\NoSelect \HasChildren) "/" "[PMail]"`)
		return goimap.CommandResponse{
			Type:    goimap.SUCCESS,
			Data:    ret,
			Message: "Success",
		}
	}

	ret = group.MatchGroup(session.Ctx.(*context.Context), basePath, template)

	return goimap.CommandResponse{
		Type:    goimap.SUCCESS,
		Data:    ret,
		Message: "Success",
	}
}

func (a action) Append(session *goimap.Session, item string) goimap.CommandResponse {
	log.Infof("%s,%s", "Append", item)
	return goimap.CommandResponse{}
}

func (a action) Select(session *goimap.Session, path string) goimap.CommandResponse {
	log.Infof("%s,%s", "Select", path)
	paths := strings.Split(path, "/")
	session.CurrentPath = strings.Trim(paths[len(paths)-1], `"`)
	_, data := group.GetGroupStatus(session.Ctx.(*context.Context), session.CurrentPath, []string{"MESSAGES", "UNSEEN", "UIDNEXT", "UIDVALIDITY"})
	ret := []string{}
	allNum := data["MESSAGES"]
	ret = append(ret, fmt.Sprintf("* %d EXISTS", allNum))
	ret = append(ret, fmt.Sprintf("* 0 RECENT"))
	unRead := data["UNSEEN"]
	ret = append(ret, fmt.Sprintf("* OK [UNSEEN %d]", unRead))
	unionID := data["UIDVALIDITY"]
	ret = append(ret, fmt.Sprintf("* OK [UIDVALIDITY %d] UID validity status", unionID))
	nextID := data["UIDNEXT"]
	ret = append(ret, fmt.Sprintf("* OK [UIDNEXT %d] Predicted next UID", nextID))
	ret = append(ret, `* FLAGS (\Answered \Flagged \Deleted \Draft \Seen)`)
	ret = append(ret, `* OK [PERMANENTFLAGS (\* \Answered \Flagged \Deleted \Draft \Seen)] Permanent flags`)

	return goimap.CommandResponse{
		Type:    goimap.SUCCESS,
		Data:    ret,
		Message: "OK [READ-WRITE] SELECT complete",
	}
}

func (a action) Store(session *goimap.Session, mailId, flags string) goimap.CommandResponse {
	log.Infof("%s,%s,%s", "Store", mailId, flags)
	return goimap.CommandResponse{}
}

func (a action) Close(session *goimap.Session) goimap.CommandResponse {
	log.Infof("%s", "Close")
	return goimap.CommandResponse{}
}

func (a action) Expunge(session *goimap.Session) goimap.CommandResponse {
	log.Infof("%s", "Expunge")
	return goimap.CommandResponse{}
}

func (a action) Examine(session *goimap.Session, path string) goimap.CommandResponse {
	log.Infof("%s,%s", "Examine", path)
	return goimap.CommandResponse{}
}

func (a action) Subscribe(session *goimap.Session, path string) goimap.CommandResponse {
	log.Infof("%s,%s", "Subscribe", path)
	return goimap.CommandResponse{}
}

func (a action) UnSubscribe(session *goimap.Session, path string) goimap.CommandResponse {
	log.Infof("%s,%s", "UnSubscribe", path)
	return goimap.CommandResponse{}
}

func (a action) LSub(session *goimap.Session, path, mailbox string) goimap.CommandResponse {
	log.Infof("%s,%s,%s", "LSub", path, mailbox)
	return goimap.CommandResponse{}
}

func (a action) Status(session *goimap.Session, mailbox string, category []string) goimap.CommandResponse {
	log.Infof("%s,%s,%+v", "Status", mailbox, category)

	category = array.Intersect([]string{"MESSAGES", "UIDNEXT", "UIDVALIDITY", "UNSEEN"}, category)
	if len(category) == 0 {
		category = []string{"MESSAGES", "UIDNEXT", "UIDVALIDITY", "UNSEEN"}
	}

	ret, _ := group.GetGroupStatus(session.Ctx.(*context.Context), mailbox, category)
	return goimap.CommandResponse{
		Type:    goimap.SUCCESS,
		Data:    []string{fmt.Sprintf(`* STATUS "%s" %s`, mailbox, ret)},
		Message: "STATUS completed",
	}

}

func (a action) Check(session *goimap.Session) goimap.CommandResponse {
	log.Infof("%s", "Check")
	return goimap.CommandResponse{}
}

func (a action) Search(session *goimap.Session, keyword, criteria string) goimap.CommandResponse {
	log.Infof("%s,%s,%s", "Search", keyword, criteria)
	return goimap.CommandResponse{}
}

func (a action) Copy(session *goimap.Session, mailId, mailBoxName string) goimap.CommandResponse {
	log.Infof("%s,%s,%s", "Copy", mailId, mailBoxName)
	return goimap.CommandResponse{}
}

func (a action) CapaBility(session *goimap.Session) goimap.CommandResponse {
	log.Infof("%s", "CapaBility")

	return goimap.CommandResponse{
		Type: goimap.SUCCESS,
		Data: []string{
			"CAPABILITY",
			"IMAP4rev1",
			"UNSELECT",
			"IDLE",
			"AUTH=LOGIN",
		},
	}

}

func (a action) IDLE(session *goimap.Session) goimap.CommandResponse {
	pools, ok := idlePool.Load(session.Account)
	if !ok {
		idlePool.Store(session.Account, []*goimap.Session{
			session,
		})
	} else {
		sPools, ok := pools.([]*goimap.Session)
		if !ok {
			idlePool.Delete(session.Account)
		} else {
			sPools = append(sPools, session)
			idlePool.Store(session.Account, sPools)
		}
	}
	return goimap.CommandResponse{}
}

func (a action) Unselect(session *goimap.Session) goimap.CommandResponse {
	log.Infof("%s", "Unselect")
	session.CurrentPath = ""
	return goimap.CommandResponse{}
}

func (a action) Noop(session *goimap.Session) goimap.CommandResponse {
	log.Infof("%s", "Noop")
	return goimap.CommandResponse{}
}

func (a action) Login(session *goimap.Session, username, pwd string) goimap.CommandResponse {
	log.WithContext(session.Ctx).Infof("%s,%s,%s", "Login", username, pwd)

	if strings.Contains(username, "@") {
		datas := strings.Split(username, "@")
		username = datas[0]
	}

	if session.Ctx == nil {
		tc := &context.Context{}
		tc.SetValue(context.LogID, id.GenLogID())
		session.Ctx = tc
	}

	var user models.User

	encodePwd := password.Encode(pwd)

	_, err := db.Instance.Where("account =? and password =? and disabled = 0", username, encodePwd).Get(&user)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.WithContext(session.Ctx.(*context.Context)).Errorf("%+v", err)
	}

	if user.ID > 0 {
		session.Status = goimap.AUTHORIZED

		session.Ctx.(*context.Context).UserID = user.ID
		session.Ctx.(*context.Context).UserName = user.Name
		session.Ctx.(*context.Context).UserAccount = user.Account

		return goimap.CommandResponse{}
	}

	return goimap.CommandResponse{
		Type:    goimap.NO,
		Message: "[AUTHENTICATIONFAILED] Invalid credentials (Failure)",
	}
}

func (a action) Logout(session *goimap.Session) goimap.CommandResponse {
	session.Status = goimap.UNAUTHORIZED

	return goimap.CommandResponse{
		Type: goimap.SUCCESS,
	}
}

func (a action) Custom(session *goimap.Session, cmd string, args string) goimap.CommandResponse {
	log.Infof("Custom  %s,%+v", cmd, args)
	return goimap.CommandResponse{}
}
