package smtp_server

import (
	"bytes"
	"database/sql"
	oerrors "errors"
	"io"
	"net"
	"net/netip"
	"strings"
	"time"

	"github.com/Jinnrry/pmail/config"
	"github.com/Jinnrry/pmail/consts"
	"github.com/Jinnrry/pmail/db"
	"github.com/Jinnrry/pmail/dto/parsemail"
	"github.com/Jinnrry/pmail/hooks"
	"github.com/Jinnrry/pmail/hooks/framework"
	"github.com/Jinnrry/pmail/listen/imap_server"
	"github.com/Jinnrry/pmail/models"
	"github.com/Jinnrry/pmail/services/rule"
	"github.com/Jinnrry/pmail/utils/array"
	"github.com/Jinnrry/pmail/utils/async"
	"github.com/Jinnrry/pmail/utils/context"
	"github.com/Jinnrry/pmail/utils/send"
	"github.com/mileusna/spf"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
)

// DropUnknownRecipientEmails 是代码级别的功能开关
// 当设置为 true 时，发给不存在用户的邮件将被直接丢弃，不会保存到数据库
// 这可以有效防止扫描器产生的垃圾邮件被存入第一个用户（管理员）的邮箱
// 注意：此开关只能通过修改代码来更改，不暴露给用户配置
const DropUnknownRecipientEmails = true

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
			return oerrors.New("No Auth")
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

		return s.deliverOutgoingEmail(email, send.Send)

	} else {
		// 收件
		var dkimStatus, SPFStatus bool

		// DKIM校验
		dkimStatus = parsemail.Check(ctx, bytes.NewReader(emailData))

		SPFStatus = spfCheck(s.RemoteAddress.String(), email.Sender, email.Sender.EmailAddress)

		_, formDomain := email.From.GetDomainAccount()
		spoofed := array.InArray(formDomain, config.Instance.Domains) && SPFStatus == false
		if spoofed {
			dkimStatus = false
		}
		email.Authentication = parsemail.NewEmailAuthentication(SPFStatus, dkimStatus)

		log.WithContext(ctx).Debugf("开始执行插件ReceiveParseAfter！")
		for _, hook := range hooks.HookList {
			if hook == nil {
				continue
			}
			email.Authentication = parsemail.NewEmailAuthentication(SPFStatus, dkimStatus)
			hook.ReceiveParseAfter(ctx, email)
		}
		log.WithContext(ctx).Debugf("开始执行插件ReceiveParseAfter！End")
		email.Authentication = parsemail.NewEmailAuthentication(SPFStatus, dkimStatus)
		// 伪造邮件状态必须覆盖插件分类结果，保持原有处理优先级。
		if spoofed {
			email.Status = 3
		}

		users, dbEmail, saveErr := saveEmail(ctx, len(emailData), email, 0, 0, s.To, SPFStatus, dkimStatus)
		if saveErr != nil {
			log.WithContext(ctx).Errorf("CRITICAL audit_event=initial_persist_failed direction=inbound msg_id=%q error=%v", email.MsgID, saveErr)
			return localPersistenceSMTPError()
		}

		if email.MessageId > 0 {
			log.WithContext(ctx).Debugf("开始执行邮件规则！")
			for _, user := range users {
				// 执行邮件规则
				rs := rule.GetAllRules(ctx, user.ID)
				for _, r := range rs {
					if rule.MatchRule(ctx, r, email) {
						rule.DoRule(ctx, r, email, user, emailData)
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

		// IDLE命令通知
		for _, user := range users {
			imap_server.IdleNotice(ctx, user.ID, dbEmail)
		}

	}

	return nil
}

func saveEmail(ctx *context.Context, size int, email *parsemail.Email, sendUserID int, emailType int, reallyTo []string, SPFStatus, dkimStatus bool) ([]*models.User, *models.Email, error) {
	if email == nil {
		return nil, nil, oerrors.New("email is missing")
	}
	if email.From == nil {
		return nil, nil, oerrors.New("email sender is missing")
	}
	if emailType == int(consts.EmailTypeReceive) && email.Authentication != nil {
		SPFStatus = email.Authentication.SPFPassed
		dkimStatus = email.Authentication.DKIMPassed
	}

	users, drop, err := resolveIncomingUsers(ctx, email, emailType, reallyTo, SPFStatus, dkimStatus)
	if err != nil || drop {
		return users, nil, err
	}

	msgID := email.MsgID
	if msgID == "" {
		msgID = parsemail.GenerateMsgID(config.Instance.Domain)
	}
	if email.Size == 0 {
		email.Size = size
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
		SPFCheck:     boolInt8(SPFStatus),
		DKIMCheck:    boolInt8(dkimStatus),
		SendUserID:   sendUserID,
		SendDate:     time.Now(),
		Status:       cast.ToInt8(email.Status),
		CreateTime:   time.Now(),
		CronSendTime: time.Now(),
		MsgID:        msgID,
	}

	var userIDs []int
	if emailType == int(consts.EmailTypeReceive) {
		for _, user := range users {
			userIDs = append(userIDs, user.ID)
		}
	} else {
		modelEmail.Status = consts.EmailStatusDeliveryPending
		modelEmail.Error = sql.NullString{String: deliveryPendingError, Valid: true}
		userIDs = []int{sendUserID}
	}

	log.WithContext(ctx).Debugf("开始入库！")
	saved, err := persistInitialAudit(ctx, modelEmail, userIDs)
	if err != nil {
		return users, nil, err
	}
	email.MessageId = cast.ToInt64(saved.Id)
	email.MsgID = saved.MsgID
	return users, saved, nil
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
