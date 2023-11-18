package pop3_server

import (
	"database/sql"
	"github.com/Jinnrry/gopop"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"pmail/db"
	"pmail/models"
	"pmail/services/detail"
	"pmail/utils/array"
	"pmail/utils/context"
	"pmail/utils/errors"
	"pmail/utils/id"
	"pmail/utils/password"
	"strings"
)

type action struct {
}

func (a action) Custom(session *gopop.Session, cmd string, args []string) ([]string, error) {
	if session.Ctx == nil {
		tc := &context.Context{}
		tc.SetValue(context.LogID, id.GenLogID())
		session.Ctx = tc
	}

	log.WithContext(session.Ctx).Warnf("not supported cmd request! cmd:%s args:%v", cmd, args)
	return nil, nil
}

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

	return ret, nil
}

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

func (a action) Pass(session *gopop.Session, pwd string) error {
	if session.Ctx == nil {
		tc := &context.Context{}
		tc.SetValue(context.LogID, id.GenLogID())
		session.Ctx = tc
	}

	log.WithContext(session.Ctx).Debugf("POP3 PASS %s , User:%s", pwd, session.User)

	var user models.User

	encodePwd := password.Encode(pwd)

	err := db.Instance.Get(&user, db.WithContext(session.Ctx.(*context.Context), "select * from user where account =? and password =?"), session.User, encodePwd)
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

	err := db.Instance.Get(&user, db.WithContext(session.Ctx.(*context.Context), "select * from user where account =? "), username)
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

type statInfo struct {
	Num  int64 `json:"num"`
	Size int64 `json:"size"`
}

func (a action) Stat(session *gopop.Session) (msgNum, msgSize int64, err error) {
	log.WithContext(session.Ctx).Debugf("POP3 CMD: STAT")

	var si statInfo
	err = db.Instance.Get(&si, db.WithContext(session.Ctx.(*context.Context), "select count(1) as `num`, sum(length(text)+length(html)) as `size` from email"))
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.WithContext(session.Ctx.(*context.Context)).Errorf("%+v", err)
		err = nil
		log.WithContext(session.Ctx).Debugf("POP3 STAT RETURT :0,0")
		return 0, 0, nil
	}
	log.WithContext(session.Ctx).Debugf("POP3 STAT RETURT : %d,%d", si.Num, si.Size)

	return si.Num, si.Size, nil
}

func (a action) Uidl(session *gopop.Session, msg string) ([]gopop.UidlItem, error) {
	log.WithContext(session.Ctx).Debugf("POP3 CMD: UIDL ,Args:%s", msg)

	reqId := cast.ToInt64(msg)
	if reqId > 0 {
		return []gopop.UidlItem{
			{
				Id:      reqId,
				UnionId: msg,
			},
		}, nil
	}

	var res []listItem

	var err error
	var ssql string

	ssql = db.WithContext(session.Ctx.(*context.Context), "SELECT id FROM email")
	err = db.Instance.Select(&res, ssql)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.WithContext(session.Ctx.(*context.Context)).Errorf("SQL:%s  Error: %+v", ssql, err)
		err = nil
		return []gopop.UidlItem{}, nil
	}
	ret := []gopop.UidlItem{}
	for _, re := range res {
		ret = append(ret, gopop.UidlItem{
			Id:      re.Id,
			UnionId: cast.ToString(re.Id),
		})
	}
	return ret, nil
}

type listItem struct {
	Id   int64 `json:"id"`
	Size int64 `json:"size"`
}

func (a action) List(session *gopop.Session, msg string) ([]gopop.MailInfo, error) {
	log.WithContext(session.Ctx).Debugf("POP3 CMD: LIST ,Args:%s", msg)
	var res []listItem
	var listId int64
	if msg != "" {
		listId = cast.ToInt64(msg)
		if listId == 0 {
			return nil, errors.New("params error")
		}
	}
	var err error
	var ssql string

	if listId != 0 {
		ssql = db.WithContext(session.Ctx.(*context.Context), "SELECT id, ifnull(LENGTH(TEXT) , 0) + ifnull(LENGTH(html) , 0) AS `size` FROM email where id =?")
		err = db.Instance.Select(&res, ssql, listId)
	} else {
		ssql = db.WithContext(session.Ctx.(*context.Context), "SELECT id, ifnull(LENGTH(TEXT) , 0) + ifnull(LENGTH(html) , 0) AS `size` FROM email")
		err = db.Instance.Select(&res, ssql)
	}

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.WithContext(session.Ctx.(*context.Context)).Errorf("SQL:%s  Error: %+v", ssql, err)
		err = nil
		return []gopop.MailInfo{}, nil
	}
	ret := []gopop.MailInfo{}
	for _, re := range res {
		ret = append(ret, gopop.MailInfo{
			Id:   re.Id,
			Size: re.Size,
		})
	}
	return ret, nil
}

func (a action) Retr(session *gopop.Session, id int64) (string, int64, error) {
	log.WithContext(session.Ctx).Debugf("POP3 CMD: RETR ,Args:%d", id)
	email, err := detail.GetEmailDetail(session.Ctx.(*context.Context), cast.ToInt(id), false)
	if err != nil {
		log.WithContext(session.Ctx.(*context.Context)).Errorf("%+v", err)
		return "", 0, errors.New("server error")
	}

	ret := email.ToTransObj().BuildBytes(session.Ctx.(*context.Context), false)
	return string(ret), cast.ToInt64(len(ret)), nil

}

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

	ret := email.ToTransObj().BuildBytes(session.Ctx.(*context.Context), false)
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

	return array.Join(res[0:headerEndLine+n+1], "\n"), nil

}

func (a action) Noop(session *gopop.Session) error {
	log.WithContext(session.Ctx).Debugf("POP3 CMD: NOOP ")
	return nil
}

func (a action) Quit(session *gopop.Session) error {
	log.WithContext(session.Ctx).Debugf("POP3 CMD: QUIT ")
	if len(session.DeleteIds) > 0 {

		_, err := db.Instance.Exec(db.WithContext(session.Ctx.(*context.Context), "DELETE FROM email WHERE id in ?"), session.DeleteIds)
		if err != nil {
			log.WithContext(session.Ctx.(*context.Context)).Errorf("%+v", err)
		}
	}

	return nil
}
