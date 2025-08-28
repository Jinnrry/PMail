package setup

import (
	"strings"

	"github.com/Jinnrry/pmail/config"
	"github.com/Jinnrry/pmail/utils/array"
	"github.com/Jinnrry/pmail/utils/errors"
)

func GetDomainSettings() (string, string, []string, int, int, int, int, int, int, error) {
	configData, err := config.ReadConfig()
	if err != nil {
		return "", "", []string{}, 0, 0, 0, 0, 0, 0, errors.Wrap(err)
	}

	smtpPort := configData.SMTPPort
	if smtpPort == 0 {
		smtpPort = 25
	}
	imapPort := configData.IMAPPort
	if imapPort == 0 {
		imapPort = 993
	}
	pop3Port := configData.POP3Port
	if pop3Port == 0 {
		pop3Port = 110
	}
	smtpsPort := configData.SMTPSPort
	if smtpsPort == 0 {
		smtpsPort = 465
	}
	imapsPort := configData.IMAPSPort
	if imapsPort == 0 {
		imapsPort = 993
	}
	pop3sPort := configData.POP3SPort
	if pop3sPort == 0 {
		pop3sPort = 995
	}

	return configData.Domain, configData.WebDomain, array.Difference(configData.Domains, []string{configData.Domain}), smtpPort, imapPort, pop3Port, smtpsPort, imapsPort, pop3sPort, nil
}

func SetDomainSettings(smtpDomain, webDomain, multiDomains string, smtpPort, imapPort, pop3Port, smtpsPort, imapsPort, pop3sPort int) error {
	configData, err := config.ReadConfig()
	if err != nil {
		return errors.Wrap(err)
	}

	if smtpDomain == "" {
		return errors.New("domain must not empty!")
	}

	if webDomain == "" {
		return errors.New("web domain must not empty!")
	}

	configData.Domains = []string{}

	if multiDomains != "" {
		domains := strings.Split(multiDomains, ",")
		configData.Domains = domains
	}

	if !array.InArray(smtpDomain, configData.Domains) {
		configData.Domains = append(configData.Domains, smtpDomain)
	}

	configData.Domain = smtpDomain
	configData.WebDomain = webDomain
	configData.SMTPPort = smtpPort
	configData.IMAPPort = imapPort
	configData.POP3Port = pop3Port
	configData.SMTPSPort = smtpsPort
	configData.IMAPSPort = imapsPort
	configData.POP3SPort = pop3sPort

	// 检查域名是否指向本机 todo

	err = config.WriteConfig(configData)
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}
