package pop3_server

import (
	"database/sql"
	"github.com/Jinnrry/gopop"
	"github.com/Jinnrry/pmail/consts"
	"github.com/Jinnrry/pmail/db"
	"github.com/Jinnrry/pmail/dto"
	"github.com/Jinnrry/pmail/dto/parsemail"
	"github.com/Jinnrry/pmail/models"
	"github.com/Jinnrry/pmail/services/del_email"
	"github.com/Jinnrry/pmail/services/detail"
	"github.com/Jinnrry/pmail/services/list"
	"github.com/Jinnrry/pmail/utils/array"
	"github.com/Jinnrry/pmail/utils/context"
	"github.com/Jinnrry/pmail/utils/errors"
	"github.com/Jinnrry/pmail/utils/id"
	"github.com/Jinnrry/pmail/utils/password"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"strings"
)

type action struct {
}

// Custom 非标准命令
func (a action) Custom(session *gopop.Session, cmd string, args []string) ([]string, error) {
	if session.Ctx == nil {
		tc := &context.Context{}
		tc.SetValue(context.LogID, id.GenLogID())
		session.Ctx = tc
	}

	log.WithContext(session.Ctx).Warnf("not supported cmd request! cmd:%s args:%v", cmd, args)
	return nil, nil
}

// Capa 说明服务端支持的命令列表
func (a action) Capa(session *gopop.Session) ([]string, error) {
	if session.Ctx == nil {
		tc := &context.Context{}
		tc.SetValue(context.LogID, id.GenLogID())
		session.Ctx = tc
	}

	if session.InTls {
		log.WithContext(session.Ctx).Debugf("POP3 CMD: CAPA With Tls")
	} else {
		log.WithContext(session.Ctx).Debugf("POP3 CMD: CAPA Without Tls")
	}

	ret := []string{
		"USER",
		"PASS",
		"TOP",
		"APOP",
		"STAT",
		"UIDL",
		"LIST",
		"RETR",
		"DELE",
		"REST",
		"NOOP",
		"QUIT",
	}
	if !session.InTls {
		ret = append(ret, "STLS")
	}

	log.WithContext(session.Ctx).Debugf("CAPA \n %+v", ret)

	return ret, nil
}

// User 提交登陆的用户名
func (a action) User(session *gopop.Session, username string) error {
	if session.Ctx == nil {
		tc := &context.Context{}
		tc.SetValue(context.LogID, id.GenLogID())
		session.Ctx = tc
	}
	log.WithContext(session.Ctx).Debugf("POP3 CMD: USER, Args:%s", username)

	infos := strings.Split(username, "@")
	if len(infos) > 1 {
		username = infos[0]
	}

	log.WithContext(session.Ctx).Debugf("POP3 User %s", username)

	session.User = username
	return nil
}

// Pass 提交密码验证
func (a action) Pass(session *gopop.Session, pwd string) error {
	if session.Ctx == nil {
		tc := &context.Context{}
		tc.SetValue(context.LogID, id.GenLogID())
		session.Ctx = tc
	}

	log.WithContext(session.Ctx).Debugf("POP3 PASS %s , User:%s", pwd, session.User)

	var user models.User

	encodePwd := password.Encode(pwd)

	_, err := db.Instance.Where("account =? and password =? and disabled = 0", session.User, encodePwd).Get(&user)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.WithContext(session.Ctx.(*context.Context)).Errorf("%+v", err)
	}

	if user.ID > 0 {
		session.Status = gopop.TRANSACTION

		session.Ctx.(*context.Context).UserID = user.ID
		session.Ctx.(*context.Context).UserName = user.Name
		session.Ctx.(*context.Context).UserAccount = user.Account

		return nil
	}

	return errors.New("password error")
}

// Apop APOP登陆命令
func (a action) Apop(session *gopop.Session, username, digest string) error {
	if session.Ctx == nil {
		tc := &context.Context{}
		tc.SetValue(context.LogID, id.GenLogID())
		session.Ctx = tc
	}
	log.WithContext(session.Ctx).Debugf("POP3 CMD: APOP, Args:%s,%s", username, digest)

	infos := strings.Split(username, "@")
	if len(infos) > 1 {
		username = infos[0]
	}

	log.WithContext(session.Ctx).Debugf("POP3 APOP %s %s", username, digest)

	var user models.User

	_, err := db.Instance.Where("account =? and disabled = 0", username).Get(&user)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.WithContext(session.Ctx.(*context.Context)).Errorf("%+v", err)
	}

	if user.ID > 0 && digest == password.Md5Encode(user.Password) {
		session.User = username
		session.Status = gopop.TRANSACTION

		session.Ctx.(*context.Context).UserID = user.ID
		session.Ctx.(*context.Context).UserName = user.Name
		session.Ctx.(*context.Context).UserAccount = user.Account

		return nil
	}

	return errors.New("password error")

}

// Stat 查询邮件数量
func (a action) Stat(session *gopop.Session) (msgNum, msgSize int64, err error) {
	log.WithContext(session.Ctx).Debugf("POP3 CMD: STAT")

	num, size := list.Stat(session.Ctx.(*context.Context))
	log.WithContext(session.Ctx).Debugf("POP3 STAT RETURT : %d,%d", num, size)
	return num, size, nil
}

// Uidl 查询某封邮件的唯一标志符
func (a action) Uidl(session *gopop.Session, msg string) ([]gopop.UidlItem, error) {
	log.WithContext(session.Ctx).Debugf("POP3 CMD: UIDL ,Args:%s", msg)

	reqId := cast.ToInt64(msg)
	if reqId > 0 {
		log.WithContext(session.Ctx).Debugf("Uidl \n %+v", reqId)
		return []gopop.UidlItem{
			{
				Id:      reqId,
				UnionId: msg,
			},
		}, nil
	}

	var res []listItem

	emailList, _ := list.GetEmailList(session.Ctx.(*context.Context), dto.SearchTag{Type: consts.EmailTypeReceive, Status: -1, GroupId: -1}, "", true, 0, 99999)
	for _, info := range emailList {
		res = append(res, listItem{
			Id:   cast.ToInt64(info.Id),
			Size: cast.ToInt64(info.Size),
		})
	}
	ret := []gopop.UidlItem{}
	for _, re := range res {
		ret = append(ret, gopop.UidlItem{
			Id:      re.Id,
			UnionId: cast.ToString(re.Id),
		})
	}

	log.WithContext(session.Ctx).Debugf("Uidl \n %+v", ret)
	return ret, nil
}

type listItem struct {
	Id   int64 `json:"id"`
	Size int64 `json:"size"`
}

// List 邮件列表
func (a action) List(session *gopop.Session, msg string) ([]gopop.MailInfo, error) {
	log.WithContext(session.Ctx).Debugf("POP3 CMD: LIST ,Args:%s", msg)
	var res []listItem
	var listId int
	if msg != "" {
		listId = cast.ToInt(msg)
		if listId == 0 {
			return nil, errors.New("params error")
		}
	}

	if listId != 0 {
		info, err := detail.GetEmailDetail(session.Ctx.(*context.Context), listId, false)
		if err != nil {
			return nil, err
		}
		item := listItem{
			Id:   cast.ToInt64(info.Id),
			Size: cast.ToInt64(info.Size),
		}
		if item.Size == 0 {
			item.Size = 9999
		}
		res = append(res, item)
	} else {
		emailList, _ := list.GetEmailList(session.Ctx.(*context.Context), dto.SearchTag{Type: consts.EmailTypeReceive, Status: -1, GroupId: -1}, "", true, 0, 99999)
		for _, info := range emailList {
			item := listItem{
				Id:   cast.ToInt64(info.Id),
				Size: cast.ToInt64(info.Size),
			}
			if item.Size == 0 {
				item.Size = 9999
			}
			res = append(res, item)
		}
	}
	ret := []gopop.MailInfo{}
	for _, re := range res {
		ret = append(ret, gopop.MailInfo{
			Id:   re.Id,
			Size: re.Size,
		})
	}

	log.WithContext(session.Ctx).Debugf("List \n %+v", ret)
	return ret, nil
}

// Retr 获取邮件详情
func (a action) Retr(session *gopop.Session, id int64) (string, int64, error) {
	log.WithContext(session.Ctx).Debugf("POP3 CMD: RETR ,Args:%d", id)
	email, err := detail.GetEmailDetail(session.Ctx.(*context.Context), cast.ToInt(id), false)
	if err != nil {
		log.WithContext(session.Ctx.(*context.Context)).Errorf("%+v", err)
		return "", 0, errors.New("server error")
	}

	ret := parsemail.NewEmailFromModel(email.Email).BuildBytes(session.Ctx.(*context.Context), false)
	log.WithContext(session.Ctx).Debugf("Retr \n %+v", string(ret))
	return string(ret), cast.ToInt64(len(ret)), nil

}

// Delete 删除邮件
func (a action) Delete(session *gopop.Session, id int64) error {
	log.WithContext(session.Ctx).Debugf("POP3 CMD: DELE ,Args:%d", id)

	session.DeleteIds = append(session.DeleteIds, id)
	session.DeleteIds = array.Unique(session.DeleteIds)
	return nil
}

func (a action) Rest(session *gopop.Session) error {
	log.WithContext(session.Ctx).Debugf("POP3 CMD: REST ")
	session.DeleteIds = []int64{}
	return nil
}

func (a action) Top(session *gopop.Session, id int64, n int) (string, error) {
	log.WithContext(session.Ctx).Debugf("POP3 CMD: TOP %d %d", id, n)
	email, err := detail.GetEmailDetail(session.Ctx.(*context.Context), cast.ToInt(id), false)
	if err != nil {
		log.WithContext(session.Ctx.(*context.Context)).Errorf("%+v", err)
		return "", errors.New("server error")
	}

	ret := parsemail.NewEmailFromModel(email.Email).BuildBytes(session.Ctx.(*context.Context), false)
	res := strings.Split(string(ret), "\n")
	headerEndLine := len(res) - 1
	for i, re := range res {
		if re == "\r" {
			headerEndLine = i
			break
		}
	}
	if len(res) <= headerEndLine+n+1 {
		return string(ret), nil
	}

	lines := array.Join(res[0:headerEndLine+n+1], "\n")
	log.WithContext(session.Ctx).Debugf("Top \n %+v", lines)
	return lines, nil

}

func (a action) Noop(session *gopop.Session) error {
	log.WithContext(session.Ctx).Debugf("POP3 CMD: NOOP ")
	return nil
}

func (a action) Quit(session *gopop.Session) error {
	log.WithContext(session.Ctx).Debugf("POP3 CMD: QUIT ")
	if len(session.DeleteIds) > 0 {
		del_email.DelEmail(session.Ctx.(*context.Context), session.DeleteIds, false)
	}

	return nil
}
