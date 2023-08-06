package setup

import (
	"pmail/utils/errors"
)

func GetDomainSettings() (string, string, error) {
	configData, err := readConfig()
	if err != nil {
		return "", "", errors.Wrap(err)
	}

	return configData.Domain, configData.WebDomain, nil
}

func SetDomainSettings(smtpDomain, webDomain string) error {
	configData, err := readConfig()
	if err != nil {
		return errors.Wrap(err)
	}

	configData.Domain = smtpDomain
	configData.WebDomain = webDomain

	// 检查域名是否指向本机 todo

	err = writeConfig(configData)
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}
