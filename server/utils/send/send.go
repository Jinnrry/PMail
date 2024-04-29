package send

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	log "github.com/sirupsen/logrus"
	"net"
	"pmail/dto/parsemail"
	"pmail/utils/array"
	"pmail/utils/async"
	"pmail/utils/consts"
	"pmail/utils/context"
	"pmail/utils/smtp"
	"strings"
	"sync"
)

type mxDomain struct {
	domain string
	mxHost string
}

// Forward 转发邮件
func Forward(ctx *context.Context, e *parsemail.Email, forwardAddress string) error {

	log.WithContext(ctx).Debugf("开始转发邮件")
	b := e.ForwardBuildBytes(ctx, forwardAddress)

	var to []*parsemail.User
	to = []*parsemail.User{
		{EmailAddress: forwardAddress},
	}

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
					domain: "smtp." + args[1],
					mxHost: "smtp." + args[1],
				}
				if err != nil {
					log.WithContext(ctx).Errorf(s.EmailAddress, "域名mx记录查询失败")
				}
				if len(mxInfo) > 0 {
					address = mxDomain{
						domain: args[1],
						mxHost: mxInfo[0].Host,
					}
				}
				toByDomain[address] = append(toByDomain[address], s)
			}
		} else {
			log.WithContext(ctx).Errorf("邮箱地址解析错误！ %s", s)
			continue
		}
	}

	var errEmailAddress []string

	as := async.New(ctx)
	for domain, tos := range toByDomain {
		domain := domain
		tos := tos
		as.WaitProcess(func(p any) {
			err := smtp.SendMail("", domain.mxHost+":25", nil, e.From.EmailAddress, buildAddress(tos), b)
			if err != nil {
				log.WithContext(ctx).Warnf("SMTP Send Error! Error:%+v", err)
			} else {
				log.WithContext(ctx).Infof("SMTP Send Success !")
			}

			// 重新选取证书域名
			if err != nil {
				if certificateErr, ok := err.(*tls.CertificateVerificationError); ok {
					if hostnameErr, is := certificateErr.Err.(x509.HostnameError); is {
						if hostnameErr.Certificate != nil {
							certificateHostName := hostnameErr.Certificate.DNSNames
							err = smtp.SendMail(domainMatch(domain.domain, certificateHostName), domain.mxHost+":25", nil, e.From.EmailAddress, buildAddress(tos), b)
							if err != nil {
								log.WithContext(ctx).Warnf("SMTP Send Error! Error:%+v", err)
							} else {
								log.WithContext(ctx).Infof("SMTP Send Success !")
							}
						}
					}
				}
			}

			if err != nil {
				log.WithContext(ctx).Errorf("%v 邮件投递失败%+v", tos, err)
				for _, user := range tos {
					errEmailAddress = append(errEmailAddress, user.EmailAddress)
				}
			}
		}, nil)
	}
	as.Wait()

	if len(errEmailAddress) > 0 {
		return errors.New("以下收件人投递失败：" + array.Join(errEmailAddress, ","))
	}
	return nil
}

func Send(ctx *context.Context, e *parsemail.Email) (error, map[string]error) {

	b := e.BuildBytes(ctx, true)

	log.WithContext(ctx).Debugf("Message Infos : %s", string(b))

	var to []*parsemail.User
	to = append(append(append(to, e.To...), e.Cc...), e.Bcc...)

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
					domain: "smtp." + args[1],
					mxHost: "smtp." + args[1],
				}
				if err != nil {
					log.WithContext(ctx).Errorf(s.EmailAddress, "域名mx记录查询失败")
				}
				if len(mxInfo) > 0 {
					address = mxDomain{
						domain: args[1],
						mxHost: mxInfo[0].Host,
					}
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

			err := smtp.SendMail("", domain.mxHost+":25", nil, e.From.EmailAddress, buildAddress(tos), b)

			// 使用其他方式发送
			if err != nil {
				if errors.Is(err, smtp.NoSupportSTARTTLSError) {
					err = smtp.SendMailWithTls("", domain.mxHost+":465", nil, e.From.EmailAddress, buildAddress(tos), b)
					if err != nil {
						log.WithContext(ctx).Warnf("Unsafe! %s Server Not Support SMTPS & STARTTLS", domain.domain)
						err = smtp.SendMailUnsafe("", domain.mxHost+":25", nil, e.From.EmailAddress, buildAddress(tos), b)
					}
				}

				// 证书错误，从新选取证书发送
				if certificateErr, ok := err.(*tls.CertificateVerificationError); ok {
					if hostnameErr, is := certificateErr.Err.(x509.HostnameError); is {
						if hostnameErr.Certificate != nil {
							certificateHostName := hostnameErr.Certificate.DNSNames
							// 重新选取证书发送
							err = smtp.SendMail(domainMatch(domain.domain, certificateHostName), domain.mxHost+":25", nil, e.From.EmailAddress, buildAddress(tos), b)
						}
					}
				}
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
