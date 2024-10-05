package smtp_server

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"github.com/Jinnrry/pmail/config"
	"github.com/Jinnrry/pmail/db"
	"github.com/Jinnrry/pmail/dto/parsemail"
	"github.com/Jinnrry/pmail/hooks"
	"github.com/Jinnrry/pmail/hooks/framework"
	"github.com/Jinnrry/pmail/models"
	"github.com/Jinnrry/pmail/services/rule"
	"github.com/Jinnrry/pmail/utils/array"
	"github.com/Jinnrry/pmail/utils/async"
	"github.com/Jinnrry/pmail/utils/context"
	"github.com/Jinnrry/pmail/utils/errors"
	"github.com/Jinnrry/pmail/utils/send"
	"github.com/mileusna/spf"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"io"
	"net"
	"net/netip"
	"strings"
	"time"
	. "xorm.io/builder"
)

func (s *Session) Data(r io.Reader) error {

	ctx := s.Ctx

	log.WithContext(ctx).Debugf("收到邮件")

	emailData, err := io.ReadAll(r)
	if err != nil {
		log.WithContext(ctx).Error("邮件内容无法读取", err)
		return err
	}

	log.WithContext(ctx).Debugf("%s", string(emailData))

	log.WithContext(ctx).Debugf("开始执行插件ReceiveParseBefore！")
	for _, hook := range hooks.HookList {
		if hook == nil {
			continue
		}
		hook.ReceiveParseBefore(ctx, &emailData)
	}
	log.WithContext(ctx).Debugf("开始执行插件ReceiveParseBefore End！")

	email := parsemail.NewEmailFromReader(s.To, bytes.NewReader(emailData), len(emailData))

	if s.From != "" {
		from := parsemail.BuilderUser(s.From)
		if email.From == nil {
			email.From = from
		}
		if email.From.EmailAddress != from.EmailAddress {
			// 协议中的from和邮件内容中的from不匹配，当成垃圾邮件处理
			//log.WithContext(s.Ctx).Infof("垃圾邮件，拒信")
			//return nil
		}
	}

	// 判断是收信还是转发，只要是登陆了，都当成转发处理
	if s.Ctx.UserID > 0 {
		account, _ := email.From.GetDomainAccount()
		if account != ctx.UserAccount && !ctx.IsAdmin {
			return errors.New("No Auth")
		}

		log.WithContext(ctx).Debugf("开始执行插件SendBefore！")
		for _, hook := range hooks.HookList {
			if hook == nil {
				continue
			}
			hook.SendBefore(ctx, email)
		}
		log.WithContext(ctx).Debugf("开始执行插件SendBefore！End")

		if email == nil {
			return nil
		}

		// 转发
		_, err := saveEmail(ctx, len(emailData), email, s.Ctx.UserID, 1, nil, true, true)
		if err != nil {
			log.WithContext(ctx).Errorf("Email Save Error %v", err)
		}

		errMsg := ""
		err, sendErr := send.Send(ctx, email)

		log.WithContext(ctx).Debugf("插件执行--SendAfter")

		as3 := async.New(ctx)
		for _, hook := range hooks.HookList {
			if hook == nil {
				continue
			}
			as3.WaitProcess(func(hk any) {
				hk.(framework.EmailHook).SendAfter(ctx, email, sendErr)
			}, hook)
		}
		as3.Wait()
		log.WithContext(ctx).Debugf("插件执行--SendAfter")

		if err != nil {
			errMsg = err.Error()
			_, err := db.Instance.Exec(db.WithContext(ctx, "update email set status =2 ,error=? where id = ? "), errMsg, email.MessageId)
			if err != nil {
				log.WithContext(ctx).Errorf("sql Error :%+v", err)
			}
			_, err = db.Instance.Exec(db.WithContext(ctx, "update user_email set status =2  where email_id = ? "), email.MessageId)
			if err != nil {
				log.WithContext(ctx).Errorf("sql Error :%+v", err)
			}

		} else {
			_, err := db.Instance.Exec(db.WithContext(ctx, "update email set status =1  where id = ? "), email.MessageId)
			if err != nil {
				log.WithContext(ctx).Errorf("sql Error :%+v", err)
			}
			_, err = db.Instance.Exec(db.WithContext(ctx, "update user_email set status =1  where email_id = ? "), email.MessageId)
			if err != nil {
				log.WithContext(ctx).Errorf("sql Error :%+v", err)
			}
		}

	} else {
		// 收件

		var dkimStatus, SPFStatus bool

		// DKIM校验
		dkimStatus = parsemail.Check(bytes.NewReader(emailData))

		SPFStatus = spfCheck(s.RemoteAddress.String(), email.Sender, email.Sender.EmailAddress)

		log.WithContext(ctx).Debugf("开始执行插件ReceiveParseAfter！")
		for _, hook := range hooks.HookList {
			if hook == nil {
				continue
			}
			hook.ReceiveParseAfter(ctx, email)
		}
		log.WithContext(ctx).Debugf("开始执行插件ReceiveParseAfter！End")

		// 垃圾过滤
		if config.Instance.SpamFilterLevel == 1 && !SPFStatus && !dkimStatus {
			log.WithContext(ctx).Infoln("垃圾邮件，拒信")
			return nil
		}

		if config.Instance.SpamFilterLevel == 2 && !SPFStatus {
			log.WithContext(ctx).Infoln("垃圾邮件，拒信")
			return nil
		}

		users, _ := saveEmail(ctx, len(emailData), email, 0, 0, s.To, SPFStatus, dkimStatus)

		if email.MessageId > 0 {
			log.WithContext(ctx).Debugf("开始执行邮件规则！")
			for _, user := range users {
				// 执行邮件规则
				rs := rule.GetAllRules(ctx, user.ID)
				for _, r := range rs {
					if rule.MatchRule(ctx, r, email) {
						rule.DoRule(ctx, r, email, user)
					}
				}
			}
		}

		log.WithContext(ctx).Debugf("开始执行插件ReceiveSaveAfter！")
		var ue []*models.UserEmail
		err = db.Instance.Table(&models.UserEmail{}).Where("email_id=?", email.MessageId).Find(&ue)
		if err != nil {
			log.WithContext(ctx).Errorf("sql Error :%+v", err)
		}
		as3 := async.New(ctx)
		for _, hook := range hooks.HookList {
			if hook == nil {
				continue
			}
			as3.WaitProcess(func(hk any) {
				hk.(framework.EmailHook).ReceiveSaveAfter(ctx, email, ue)
			}, hook)
		}
		as3.Wait()
		log.WithContext(ctx).Debugf("开始执行插件ReceiveSaveAfter！End")

	}

	return nil
}

func saveEmail(ctx *context.Context, size int, email *parsemail.Email, sendUserID int, emailType int, reallyTo []string, SPFStatus, dkimStatus bool) ([]*models.User, error) {
	var dkimV, spfV int8
	if dkimStatus {
		dkimV = 1
	}
	if SPFStatus {
		spfV = 1
	}

	log.WithContext(ctx).Debugf("开始入库！")

	if email == nil {
		return nil, nil
	}

	modelEmail := models.Email{
		Type:         cast.ToInt8(emailType),
		Subject:      email.Subject,
		ReplyTo:      json2string(email.ReplyTo),
		FromName:     email.From.Name,
		FromAddress:  email.From.EmailAddress,
		To:           json2string(email.To),
		Bcc:          json2string(email.Bcc),
		Cc:           json2string(email.Cc),
		Text:         sql.NullString{String: string(email.Text), Valid: true},
		Html:         sql.NullString{String: string(email.HTML), Valid: true},
		Sender:       json2string(email.Sender),
		Attachments:  json2string(email.Attachments),
		Size:         email.Size,
		SPFCheck:     spfV,
		DKIMCheck:    dkimV,
		SendUserID:   sendUserID,
		SendDate:     time.Now(),
		Status:       cast.ToInt8(email.Status),
		CreateTime:   time.Now(),
		CronSendTime: time.Now(),
	}

	_, err := db.Instance.Insert(&modelEmail)

	if err != nil {
		log.WithContext(ctx).Errorf("db insert error:%+v", err.Error())
	}

	if modelEmail.Id > 0 {
		email.MessageId = cast.ToInt64(modelEmail.Id)
	}
	// 收信人信息
	var users []*models.User

	// 如果是收信
	if emailType == 0 {
		// 找到收信人id
		accounts := []string{}
		// 优先取smtp协议中的收件人地址
		if len(reallyTo) > 0 {
			for _, s := range reallyTo {
				account := parsemail.BuilderUser(s)
				if account != nil {
					acc, domain := account.GetDomainAccount()
					if array.InArray(domain, config.Instance.Domains) && acc != "" {
						accounts = append(accounts, acc)
					}
				}
			}
		} else {
			for _, user := range append(append(email.To, email.Cc...), email.Bcc...) {
				account, _ := user.GetDomainAccount()
				if account != "" {
					accounts = append(accounts, account)
				}
			}
		}

		where, params, _ := ToSQL(In("account", accounts))

		err = db.Instance.Table(&models.User{}).Where(where, params...).Find(&users)
		if err != nil {
			log.WithContext(ctx).Errorf("db Select error:%+v", err.Error())
		}

		if len(users) > 0 {
			for _, user := range users {
				ue := models.UserEmail{EmailID: modelEmail.Id, UserID: user.ID, Status: cast.ToInt8(email.Status)}
				_, err = db.Instance.Insert(&ue)
				if err != nil {
					log.WithContext(ctx).Errorf("db insert error:%+v", err.Error())
				}
			}
		} else {
			err = db.Instance.Table(&models.User{}).Where("is_admin=1").Find(&users)
			// 当邮件找不到收件人的时候，邮件全部丢给管理员账号
			for _, user := range users {
				ue := models.UserEmail{EmailID: modelEmail.Id, UserID: user.ID, Status: cast.ToInt8(email.Status)}
				_, err = db.Instance.Insert(&ue)
				if err != nil {
					log.WithContext(ctx).Errorf("db insert error:%+v", err.Error())
				}
			}
		}
	} else {
		ue := models.UserEmail{EmailID: modelEmail.Id, UserID: ctx.UserID}
		_, err = db.Instance.Insert(&ue)
		if err != nil {
			log.WithContext(ctx).Errorf("db insert error:%+v", err.Error())
		}
	}

	return users, nil
}

func json2string(d any) string {
	by, _ := json.Marshal(d)
	return string(by)
}

func spfCheck(remoteAddress string, sender *parsemail.User, senderString string) bool {
	//spf校验
	ipAddress, _ := netip.ParseAddrPort(remoteAddress)

	ip := net.ParseIP(ipAddress.Addr().String())
	if ip.IsPrivate() {
		return true
	}

	tmp := strings.Split(sender.EmailAddress, "@")
	if len(tmp) < 2 {
		return false
	}

	res := spf.CheckHost(ip, tmp[1], senderString, "")

	if res == spf.None || res == spf.Pass {
		// spf校验通过
		return true
	}
	return false
}
