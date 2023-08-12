package setup

import (
	"pmail/utils/errors"
)

func GetDomainSettings() (string, string, error) {
	configData, err := ReadConfig()
	if err != nil {
		return "", "", errors.Wrap(err)
	}

	return configData.Domain, configData.WebDomain, nil
}

func SetDomainSettings(smtpDomain, webDomain string) error {
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

	configData.Domain = smtpDomain
	configData.WebDomain = webDomain

	// 检查域名是否指向本机 todo

	err = WriteConfig(configData)
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}
