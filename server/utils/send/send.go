package send

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/Jinnrry/pmail/config"
	"github.com/Jinnrry/pmail/dto/parsemail"
	"github.com/Jinnrry/pmail/models"
	"github.com/Jinnrry/pmail/utils/array"
	"github.com/Jinnrry/pmail/utils/async"
	"github.com/Jinnrry/pmail/utils/consts"
	"github.com/Jinnrry/pmail/utils/context"
	"github.com/Jinnrry/pmail/utils/smtp"
	log "github.com/sirupsen/logrus"
)

type mxDomain struct {
	domain string
	mxHost string
}

// Forward 转发邮件
func Forward(ctx *context.Context, e *parsemail.Email, forwardAddress string, user *models.User) error {

	log.WithContext(ctx).Debugf("开始转发邮件")

	b := e.ForwardBuildBytes(ctx, user)

	log.WithContext(ctx).Debugf("%s", b)

	var to []*parsemail.User
	to = []*parsemail.User{
		{EmailAddress: forwardAddress},
	}

	err, _ := doSend(ctx, config.Instance.Domains[0], b, to, e.From.EmailAddress)

	return err
}

func Send(ctx *context.Context, e *parsemail.Email) (error, map[string]error) {

	_, fromDomain := e.From.GetDomainAccount()

	b := e.BuildBytes(ctx, true)

	var to []*parsemail.User
	to = append(append(append(to, e.To...), e.Cc...), e.Bcc...)

	return doSend(ctx, fromDomain, b, to, e.From.EmailAddress)

}

func doSend(ctx *context.Context, fromDomain string, data []byte, to []*parsemail.User, from string) (error, map[string]error) {

	// 按域名整理
	toByDomain := map[mxDomain][]*parsemail.User{}
	for _, s := range to {
		args := strings.Split(s.EmailAddress, "@")
		if len(args) == 2 {
			if args[1] == consts.TEST_DOMAIN {
				// 测试使用
				address := mxDomain{
					domain: "localhost",
					mxHost: "127.0.0.1",
				}
				toByDomain[address] = append(toByDomain[address], s)
			} else {
				//查询dns mx记录
				mxInfo, err := net.LookupMX(args[1])
				address := mxDomain{
					domain: args[1],
					mxHost: args[1],
				}
				if err != nil {
					log.WithContext(ctx).Errorf(s.EmailAddress, "域名mx记录查询失败，检查邮箱是否存在！")
				}
				if len(mxInfo) > 0 {
					address.mxHost = mxInfo[0].Host
				}
				toByDomain[address] = append(toByDomain[address], s)
			}
		} else {
			log.WithContext(ctx).Errorf("邮箱地址解析错误！ %s", s)
			continue
		}
	}

	var errEmailAddress []string

	errMap := sync.Map{}

	as := async.New(ctx)
	for domain, tos := range toByDomain {
		domain := domain
		tos := tos
		as.WaitProcess(func(p any) {

			smtpPort := config.Instance.SMTPPort
			if smtpPort == 0 {
				smtpPort = 25
			}
			smtpAddr := fmt.Sprintf("%s:%d", domain.mxHost, smtpPort)

			if domain.domain == "localhost" {
				err := smtp.SendMailUnsafe("", smtpAddr, nil, from, fromDomain, buildAddress(tos), data)
				if err != nil {
					log.WithContext(ctx).Errorf("send error %s", err.Error())
				}
				return
			}

			// 优先尝试配置端口，starttls方式投递
			err := smtp.SendMail("", smtpAddr, nil, from, fromDomain, buildAddress(tos), data)
			if err == nil {
				return
			}
			// 证书错误，从新选取证书发送
			var certificateErr *tls.CertificateVerificationError
			if errors.As(err, &certificateErr) {
				// 单测使用
				var hostnameErr x509.HostnameError
				if errors.As(certificateErr.Err, &hostnameErr) {
					if hostnameErr.Certificate != nil {
						certificateHostName := hostnameErr.Certificate.DNSNames
						// 重新选取证书发送
						err = smtp.SendMail(domainMatch(domain.domain, certificateHostName), smtpAddr, nil, from, fromDomain, buildAddress(tos), data)
					}
				}
			}
			if err == nil {
				return
			}
			log.WithContext(ctx).Infof("SMTP STARTTLS on %d Send Error. %s", smtpPort, err.Error())

			// 再试用587投递
			err = smtp.SendMailWithTls("", domain.mxHost+":587", nil, from, fromDomain, buildAddress(tos), data)
			if err == nil {
				return
			}
			log.WithContext(ctx).Infof("SMTPS on 587 Send Error. %s", err.Error())

			// 再次尝试465投递
			err = smtp.SendMailWithTls("", domain.mxHost+":465", nil, from, fromDomain, buildAddress(tos), data)
			if err == nil {
				return
			}
			log.WithContext(ctx).Infof("SMTPS on 465 Send Error. %s", err.Error())

			// 最后尝试非安全方式投递
			err = smtp.SendMailUnsafe("", smtpAddr, nil, from, fromDomain, buildAddress(tos), data)
			if err == nil {
				log.WithContext(ctx).Warnf("Send By Unsafe SMTP")
				return
			}

			if err != nil {
				log.WithContext(ctx).Errorf("%v 邮件投递失败%+v", tos, err)
				for _, user := range tos {
					errEmailAddress = append(errEmailAddress, user.EmailAddress)
				}
			}
			errMap.Store(domain.domain, err)
		}, nil)
	}
	as.Wait()

	orgMap := map[string]error{}
	errMap.Range(func(key, value any) bool {
		if value != nil {
			orgMap[key.(string)] = value.(error)
		} else {
			orgMap[key.(string)] = nil
		}

		return true
	})

	if len(errEmailAddress) > 0 {
		return errors.New("以下收件人投递失败：" + array.Join(errEmailAddress, ",")), orgMap
	}
	return nil, orgMap
}

func buildAddress(u []*parsemail.User) []string {
	var ret []string

	for _, user := range u {
		ret = append(ret, user.EmailAddress)

	}

	return ret
}

func domainMatch(domain string, dnsNames []string) string {
	if len(dnsNames) == 0 {
		return domain
	}

	secondMatch := ""

	for _, name := range dnsNames {
		if strings.Contains(name, "smtp") {
			secondMatch = name
		}

		if name == domain {
			return name
		}
		if strings.Contains(name, "*") {
			nameArg := strings.Split(name, ".")
			domainArg := strings.Split(domain, ".")
			match := true
			for i := 0; i < len(nameArg); i++ {
				if nameArg[len(nameArg)-1-i] == "*" {
					continue
				}
				if len(domainArg) > i {
					if nameArg[len(nameArg)-1-i] == domainArg[len(domainArg)-1-i] {
						continue
					}
				}
				match = false
				break
			}

			for i := 0; i < len(domainArg); i++ {
				if len(nameArg) > i && nameArg[len(nameArg)-1-i] == domainArg[len(domainArg)-1-i] {
					continue
				}
				if len(nameArg) > i && nameArg[len(nameArg)-1-i] == "*" {
					continue
				}

				match = false
				break
			}
			if match {
				return domain
			}
		}
	}

	if secondMatch != "" {
		return strings.ReplaceAll(secondMatch, "*.", "")
	}

	return strings.ReplaceAll(dnsNames[0], "*.", "")
}
