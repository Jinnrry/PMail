package smtp_server

import (
	"bytes"
	"encoding/json"
	"github.com/mileusna/spf"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"net/netip"
	"pmail/config"
	"pmail/db"
	"pmail/dto/parsemail"
	"pmail/hooks"
	"pmail/services/rule"
	"pmail/utils/array"
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

	as1 := async.New(ctx)
	for _, hook := range hooks.HookList {
		if hook == nil {
			continue
		}
		as1.WaitProcess(func(hk any) {
			hk.(hooks.EmailHook).ReceiveParseBefore(emailData)
		}, hook)
	}
	as1.Wait()

	log.WithContext(ctx).Infof("邮件原始内容: %s", emailData)

	email := parsemail.NewEmailFromReader(s.To, bytes.NewReader(emailData))

	if s.From != "" {
		from := parsemail.BuilderUser(s.From)
		if email.From == nil {
			email.From = from
		}
		if email.From.EmailAddress != from.EmailAddress {
			// 协议中的from和邮件内容中的from不匹配，当成垃圾邮件处理
			log.WithContext(s.Ctx).Infof("垃圾邮件，拒信")
			return nil
		}
	}

	// 判断是收信还是转发
	account, domain := email.From.GetDomainAccount()
	if array.InArray(domain, config.Instance.Domains) && s.Ctx.UserName == account {
		// 转发
		err := saveEmail(ctx, email, 1, true, true)
		if err != nil {
			log.WithContext(ctx).Errorf("Email Save Error %v", err)
		}

		send.Send(ctx, email)

	} else {
		// 收件

		var dkimStatus, SPFStatus bool

		// DKIM校验
		dkimStatus = parsemail.Check(bytes.NewReader(emailData))

		if err != nil {
			log.WithContext(ctx).Errorf("邮件内容解析失败！ Error : %v \n", err)
		}

		SPFStatus = spfCheck(s.RemoteAddress.String(), email.Sender, email.Sender.EmailAddress)

		saveEmail(ctx, email, 0, SPFStatus, dkimStatus)

		log.WithContext(ctx).Debugf("开始执行插件！")

		as2 := async.New(ctx)
		for _, hook := range hooks.HookList {
			if hook == nil {
				continue
			}
			as2.WaitProcess(func(hk any) {
				hk.(hooks.EmailHook).ReceiveParseAfter(email)
			}, hook)
		}
		as2.Wait()

		log.WithContext(ctx).Debugf("开始执行邮件规则！")
		// 执行邮件规则
		rs := rule.GetAllRules(ctx)
		for _, r := range rs {
			if rule.MatchRule(ctx, r, email) {
				rule.DoRule(ctx, r, email)
			}
		}
	}

	return nil
}

func saveEmail(ctx *context.Context, email *parsemail.Email, emailType int, SPFStatus, dkimStatus bool) error {
	var dkimV, spfV int8
	if dkimStatus {
		dkimV = 1
	}
	if SPFStatus {
		spfV = 1
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

	log.WithContext(ctx).Debugf("开始入库！")

	if email == nil {
		return nil
	}

	sql := "INSERT INTO email (type, send_date, subject, reply_to, from_name, from_address, `to`, bcc, cc, text, html, sender, attachments,spf_check, dkim_check, create_time,is_read,status,group_id) VALUES (?,?,?,?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	_, err := db.Instance.Exec(sql,
		emailType,
		email.Date,
		email.Subject,
		json2string(email.ReplyTo),
		email.From.Name,
		email.From.EmailAddress,
		json2string(email.To),
		json2string(email.Bcc),
		json2string(email.Cc),
		email.Text,
		email.HTML,
		json2string(email.Sender),
		json2string(email.Attachments),
		spfV,
		dkimV,
		time.Now(),
		email.IsRead,
		email.Status,
		email.GroupId,
	)

	if err != nil {
		log.WithContext(ctx).Println("mysql insert error:", err.Error())
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
