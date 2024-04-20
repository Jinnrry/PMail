package smtp_server

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"github.com/mileusna/spf"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"io"
	"net"
	"net/netip"
	"pmail/config"
	"pmail/db"
	"pmail/dto/parsemail"
	"pmail/hooks"
	"pmail/hooks/framework"
	"pmail/models"
	"pmail/services/rule"
	"pmail/utils/async"
	"pmail/utils/context"
	"pmail/utils/send"
	"strings"
	"time"
)

func (s *Session) Data(r io.Reader) error {

	ctx := s.Ctx

	log.WithContext(ctx).Debugf("收到邮件")

	emailData, err := io.ReadAll(r)
	if err != nil {
		log.WithContext(ctx).Error("邮件内容无法读取", err)
		return err
	}
	log.WithContext(ctx).Debugf("开始执行插件ReceiveParseBefore！")
	for _, hook := range hooks.HookList {
		if hook == nil {
			continue
		}
		hook.ReceiveParseBefore(ctx, &emailData)
	}
	log.WithContext(ctx).Debugf("开始执行插件ReceiveParseBefore End！")

	log.WithContext(ctx).Infof("邮件原始内容: %s", emailData)

	email := parsemail.NewEmailFromReader(s.To, bytes.NewReader(emailData))

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
	//account, domain := email.From.GetDomainAccount()
	if s.Ctx.UserID > 0 {
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
		err := saveEmail(ctx, email, s.Ctx.UserID, 1, true, true)
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
		} else {
			_, err := db.Instance.Exec(db.WithContext(ctx, "update email set status =1  where id = ? "), email.MessageId)
			if err != nil {
				log.WithContext(ctx).Errorf("sql Error :%+v", err)
			}
		}

	} else {
		// 收件

		var dkimStatus, SPFStatus bool

		// DKIM校验
		dkimStatus = parsemail.Check(bytes.NewReader(emailData))

		if err != nil {
			log.WithContext(ctx).Errorf("邮件内容解析失败！ Error : %v \n", err)
		}

		SPFStatus = spfCheck(s.RemoteAddress.String(), email.Sender, email.Sender.EmailAddress)

		log.WithContext(ctx).Debugf("开始执行插件ReceiveParseAfter！")
		for _, hook := range hooks.HookList {
			if hook == nil {
				continue
			}
			hook.ReceiveParseAfter(ctx, email)
		}
		log.WithContext(ctx).Debugf("开始执行插件ReceiveParseAfter！End")

		if email == nil {
			return nil
		}

		// 垃圾过滤
		if config.Instance.SpamFilterLevel == 1 && !SPFStatus && !dkimStatus {
			log.WithContext(ctx).Infoln("垃圾邮件，拒信")
			return nil
		}

		if config.Instance.SpamFilterLevel == 2 && !SPFStatus {
			log.WithContext(ctx).Infoln("垃圾邮件，拒信")
			return nil
		}

		saveEmail(ctx, email, 0, 0, SPFStatus, dkimStatus)

		if email.MessageId > 0 {
			log.WithContext(ctx).Debugf("开始执行邮件规则！")
			// 执行邮件规则
			rs := rule.GetAllRules(ctx)
			for _, r := range rs {
				if rule.MatchRule(ctx, r, email) {
					rule.DoRule(ctx, r, email)
				}
			}
		}

		log.WithContext(ctx).Debugf("开始执行插件ReceiveSaveAfter！")
		as3 := async.New(ctx)
		for _, hook := range hooks.HookList {
			if hook == nil {
				continue
			}
			as3.WaitProcess(func(hk any) {
				hk.(framework.EmailHook).ReceiveSaveAfter(ctx, email)
			}, hook)
		}
		as3.Wait()
		log.WithContext(ctx).Debugf("开始执行插件ReceiveSaveAfter！End")

	}

	return nil
}

func saveEmail(ctx *context.Context, email *parsemail.Email, sendUserID int, emailType int, SPFStatus, dkimStatus bool) error {
	var dkimV, spfV int8
	if dkimStatus {
		dkimV = 1
	}
	if SPFStatus {
		spfV = 1
	}

	log.WithContext(ctx).Debugf("开始入库！")

	if email == nil {
		return nil
	}

	modelEmail := models.Email{
		Type:        cast.ToInt8(emailType),
		GroupId:     email.GroupId,
		Subject:     email.Subject,
		ReplyTo:     json2string(email.ReplyTo),
		FromName:    email.From.Name,
		FromAddress: email.From.EmailAddress,
		To:          json2string(email.To),
		Bcc:         json2string(email.Bcc),
		Cc:          json2string(email.Cc),
		Text:        sql.NullString{String: string(email.Text), Valid: true},
		Html:        sql.NullString{String: string(email.HTML), Valid: true},
		Sender:      json2string(email.Sender),
		Attachments: json2string(email.Attachments),
		SPFCheck:    spfV,
		DKIMCheck:   dkimV,
		SendUserID:  sendUserID,
		SendDate:    time.Now(),
		Status:      cast.ToInt8(email.Status),
		CreateTime:  time.Now(),
	}

	_, err := db.Instance.Insert(&modelEmail)

	if err != nil {
		log.WithContext(ctx).Errorf("db insert error:%+v", err.Error())
	}

	if modelEmail.Id > 0 {
		email.MessageId = cast.ToInt64(modelEmail.Id)
	}

	return nil
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
