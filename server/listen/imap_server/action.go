package imap_server

import (
	"database/sql"
	errors2 "errors"
	"fmt"
	"github.com/Jinnrry/pmail/db"
	"github.com/Jinnrry/pmail/models"
	"github.com/Jinnrry/pmail/services/group"
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

func (a action) Create(session *goimap.Session, path string) error {
	log.Infof("%s,%s", "Create", path)
	return nil
}

func (a action) Delete(session *goimap.Session, path string) error {
	log.Infof("%s,%s", "Delete", path)
	return nil
}

func (a action) Rename(session *goimap.Session, oldPath, newPath string) error {
	log.Infof("%s,%s,%s", "Rename", oldPath, newPath)
	return nil
}

func (a action) List(session *goimap.Session, basePath, template string) ([]string, error) {
	log.Infof("%s,%s,%s", "List", basePath, template)
	var ret []string
	if basePath == "" && template == "" {
		ret = append(ret, `* LIST (\NoSelect \HasChildren) "/" "[PMail]`)
		return ret, nil
	}

	ret = group.MatchGroup(session.Ctx.(*context.Context), basePath, template)

	return ret, nil
}

func (a action) Append(session *goimap.Session, item string) error {
	log.Infof("%s,%s", "Append", item)
	return nil
}

func (a action) Select(session *goimap.Session, path string) ([]string, error) {
	log.Infof("%s,%s", "Select", path)
	paths := strings.Split(path, "/")
	session.CurrentPath = paths[len(paths)-1]
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
	ret = append(ret, `$$NUM OK [READ-WRITE] SELECT complete`)

	return ret, nil
}

func (a action) Fetch(session *goimap.Session, mailIds, dataNames string) (string, error) {
	log.Infof("%s,%s,%s", "Fetch", mailIds, dataNames)
	return "", nil
}

func (a action) Store(session *goimap.Session, mailId, flags string) error {
	log.Infof("%s,%s,%s", "Store", mailId, flags)
	return nil
}

func (a action) Close(session *goimap.Session) error {
	log.Infof("%s", "Close")
	return nil
}

func (a action) Expunge(session *goimap.Session) error {
	log.Infof("%s", "Expunge")
	return nil
}

func (a action) Examine(session *goimap.Session, path string) error {
	log.Infof("%s,%s", "Examine", path)
	return nil
}

func (a action) Subscribe(session *goimap.Session, path string) error {
	log.Infof("%s,%s", "Subscribe", path)
	return nil
}

func (a action) UnSubscribe(session *goimap.Session, path string) error {
	log.Infof("%s,%s", "UnSubscribe", path)
	return nil
}

func (a action) LSub(session *goimap.Session, path, mailbox string) ([]string, error) {
	log.Infof("%s,%s,%s", "LSub", path, mailbox)
	return nil, nil
}

func (a action) Status(session *goimap.Session, mailbox string, category []string) (string, error) {
	log.Infof("%s,%s,%+v", "Status", mailbox, category)
	ret, _ := group.GetGroupStatus(session.Ctx.(*context.Context), mailbox, category)
	return fmt.Sprintf(`* STATUS "%s" %s`, mailbox, ret), nil
}

func (a action) Check(session *goimap.Session) error {
	log.Infof("%s", "Check")
	return nil
}

func (a action) Search(session *goimap.Session, keyword, criteria string) (string, error) {
	log.Infof("%s,%s,%s", "Search", keyword, criteria)
	return "", nil
}

func (a action) Copy(session *goimap.Session, mailId, mailBoxName string) error {
	log.Infof("%s,%s,%s", "Copy", mailId, mailBoxName)
	return nil
}

func (a action) CapaBility(session *goimap.Session) ([]string, error) {
	log.Infof("%s", "CapaBility")
	return []string{
		"CAPABILITY",
		"IMAP4rev1",
		"UNSELECT",
		"IDLE",
		"AUTH=PLAIN",
		"AUTH=LOGIN",
	}, nil
}

func (a action) IDLE(session *goimap.Session) error {
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
	return nil
}

func (a action) Unselect(session *goimap.Session) error {
	log.Infof("%s", "Unselect")
	session.CurrentPath = ""
	return nil
}

func (a action) Noop(session *goimap.Session) error {
	log.Infof("%s", "Noop")
	return nil
}

func (a action) Login(session *goimap.Session, username, pwd string) error {
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

		return nil
	}

	return errors2.New("password error")
}

func (a action) Logout(session *goimap.Session) error {
	session.Status = goimap.UNAUTHORIZED
	if session.Conn != nil {
		_ = session.Conn.Close()
	}
	return nil
}

func (a action) Custom(session *goimap.Session, cmd string, args string) ([]string, error) {
	log.Infof("Custom  %s,%+v", cmd, args)
	return nil, nil
}
