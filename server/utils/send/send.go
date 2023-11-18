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
	"pmail/utils/context"
	"pmail/utils/smtp"
	"strings"
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
		} else {
			log.WithContext(ctx).Errorf("邮箱地址解析错误！ %s", s)
			continue
		}
	}

	var errEmailAddress []string

	errMap := map[string]error{}

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
			errMap[domain.domain] = err
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

	var to []*parsemail.User
	to = append(append(append(to, e.To...), e.Cc...), e.Bcc...)

	// 按域名整理
	toByDomain := map[mxDomain][]*parsemail.User{}
	for _, s := range to {
		args := strings.Split(s.EmailAddress, "@")
		if len(args) == 2 {
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
		} else {
			log.WithContext(ctx).Errorf("邮箱地址解析错误！ %s", s)
			continue
		}
	}

	var errEmailAddress []string

	errMap := map[string]error{}

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
							// smtps发送失败，尝试smtp
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
			errMap[domain.domain] = err
		}, nil)
	}
	as.Wait()

	if len(errEmailAddress) > 0 {
		return errors.New("以下收件人投递失败：" + array.Join(errEmailAddress, ",")), errMap
	}
	return nil, errMap

}

func buildAddress(u []*parsemail.User) []string {
	var ret []string

	for _, user := range u {
		ret = append(ret, user.EmailAddress)

	}

	return ret
}

func domainMatch(domain string, dnsNames []string) string {
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
