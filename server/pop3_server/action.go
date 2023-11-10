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
)

type action struct {
}

func (a action) User(ctx *gopop.Data, username string) error {
	if ctx.Ctx == nil {
		tc := &context.Context{}
		tc.SetValue(context.LogID, id.GenLogID())
		ctx.Ctx = tc
	}

	log.WithContext(ctx.Ctx).Debugf("POP3 User %s", username)

	ctx.User = username
	return nil
}

func (a action) Pass(ctx *gopop.Data, pwd string) error {
	if ctx.Ctx == nil {
		tc := &context.Context{}
		tc.SetValue(context.LogID, id.GenLogID())
		ctx.Ctx = tc
	}

	log.WithContext(ctx.Ctx).Debugf("POP3 PASS %s", pwd)

	var user models.User

	encodePwd := password.Encode(pwd)

	err := db.Instance.Get(&user, db.WithContext(ctx.Ctx.(*context.Context), "select * from user where account =? and password =?"), ctx.User, encodePwd)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.WithContext(ctx.Ctx.(*context.Context)).Errorf("%+v", err)
	}

	if user.ID > 0 {
		ctx.Status = gopop.TRANSACTION

		ctx.Ctx.(*context.Context).UserID = user.ID
		ctx.Ctx.(*context.Context).UserName = user.Name
		ctx.Ctx.(*context.Context).UserAccount = user.Account

		return nil
	}

	return errors.New("password error")
}

func (a action) Apop(ctx *gopop.Data, username, digest string) error {
	if ctx.Ctx == nil {
		tc := &context.Context{}
		tc.SetValue(context.LogID, id.GenLogID())
		ctx.Ctx = tc
	}

	log.WithContext(ctx.Ctx).Debugf("POP3 APOP %s %s", username, digest)

	var user models.User

	err := db.Instance.Get(&user, db.WithContext(ctx.Ctx.(*context.Context), "select * from user where account =? "), username)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.WithContext(ctx.Ctx.(*context.Context)).Errorf("%+v", err)
	}

	if user.ID > 0 && digest == password.Md5Encode(user.Password) {
		ctx.User = username
		ctx.Status = gopop.TRANSACTION

		ctx.Ctx.(*context.Context).UserID = user.ID
		ctx.Ctx.(*context.Context).UserName = user.Name
		ctx.Ctx.(*context.Context).UserAccount = user.Account

		return nil
	}

	return errors.New("password error")

}

type statInfo struct {
	Num  int64 `json:"num"`
	Size int64 `json:"size"`
}

func (a action) Stat(ctx *gopop.Data) (msgNum, msgSize int64, err error) {

	var si statInfo
	err = db.Instance.Get(&si, db.WithContext(ctx.Ctx.(*context.Context), "select count(1) as `num`, sum(length(text)+length(html)) as `size` from email"))
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.WithContext(ctx.Ctx.(*context.Context)).Errorf("%+v", err)
		err = nil
		log.WithContext(ctx.Ctx).Debugf("POP3 STAT RETURT :0,0")
		return 0, 0, nil
	}
	log.WithContext(ctx.Ctx).Debugf("POP3 STAT RETURT : %d,%d", si.Num, si.Size)

	return si.Num, si.Size, nil
}

func (a action) Uidl(ctx *gopop.Data, id int64) (string, error) {
	log.WithContext(ctx.Ctx).Debugf("POP3 Uidl RETURT : %d", id)

	return cast.ToString(id), nil
}

type listItem struct {
	Id   int64 `json:"id"`
	Size int64 `json:"size"`
}

func (a action) List(ctx *gopop.Data, msg string) ([]gopop.MailInfo, error) {
	var res []listItem
	var id int64
	if msg != "" {
		id = cast.ToInt64(msg)
		if id == 0 {
			return nil, errors.New("params error")
		}
	}
	var err error
	if id != 0 {
		err = db.Instance.Select(&res, db.WithContext(ctx.Ctx.(*context.Context), "select id, length(text)+length(html) as `size` from email where id =?"), id)
	} else {
		err = db.Instance.Select(&res, db.WithContext(ctx.Ctx.(*context.Context), "select id, length(text)+length(html) as `size` from email"))
	}

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.WithContext(ctx.Ctx.(*context.Context)).Errorf("%+v", err)
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

func (a action) Retr(ctx *gopop.Data, id int64) (string, int64, error) {

	email, err := detail.GetEmailDetail(ctx.Ctx.(*context.Context), cast.ToInt(id), false)
	if err != nil {
		log.WithContext(ctx.Ctx.(*context.Context)).Errorf("%+v", err)
		return "", 0, errors.New("server error")
	}

	ret := email.ToTransObj().BuildBytes(ctx.Ctx.(*context.Context), false)
	return string(ret), cast.ToInt64(len(ret)), nil

}

func (a action) Delete(ctx *gopop.Data, id int64) error {
	ctx.DeleteIds = append(ctx.DeleteIds, id)
	ctx.DeleteIds = array.Unique(ctx.DeleteIds)
	return nil
}

func (a action) Rest(ctx *gopop.Data) error {
	ctx.DeleteIds = []int64{}
	return nil
}

func (a action) Top(ctx *gopop.Data, id int64, n int) (string, error) {
	//email, err := detail.GetEmailDetail(ctx.Ctx.(*context.Context), cast.ToInt(id), false)
	//if err != nil {
	//	log.WithContext(ctx.Ctx.(*context.Context)).Errorf("%+v", err)
	//	return "", errors.New("server error")
	//}
	//
	//ret := email.ToTransObj().BuilderHeaders(ctx.Ctx.(*context.Context))
	//return string(ret), nil

	return "", errors.New("not supported")
}

func (a action) Noop(ctx *gopop.Data) error {
	return nil
}

func (a action) Quit(ctx *gopop.Data) error {
	if len(ctx.DeleteIds) > 0 {

		_, err := db.Instance.Exec(db.WithContext(ctx.Ctx.(*context.Context), "DELETE FROM email WHERE id in ?"), ctx.DeleteIds)
		if err != nil {
			log.WithContext(ctx.Ctx.(*context.Context)).Errorf("%+v", err)
		}
	}

	return nil
}
