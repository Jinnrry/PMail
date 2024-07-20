package setup

import (
	"github.com/Jinnrry/pmail/utils/array"
	"github.com/Jinnrry/pmail/utils/errors"
	"strings"
)

func GetDomainSettings() (string, string, []string, error) {
	configData, err := ReadConfig()
	if err != nil {
		return "", "", []string{}, errors.Wrap(err)
	}

	return configData.Domain, configData.WebDomain, array.Difference(configData.Domains, []string{configData.Domain}), nil
}

func SetDomainSettings(smtpDomain, webDomain, multiDomains string) error {
	configData, err := ReadConfig()
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

	// 检查域名是否指向本机 todo

	err = WriteConfig(configData)
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}
