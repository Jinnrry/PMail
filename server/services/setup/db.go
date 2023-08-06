package setup

import (
	"encoding/json"
	"os"
	"pmail/config"
	"pmail/utils/array"
	"pmail/utils/errors"
	"pmail/utils/file"
)

func GetDatabaseSettings() (string, string, error) {
	configData, err := readConfig()
	if err != nil {
		return "", "", errors.Wrap(err)
	}

	return configData.DbType, configData.DbDSN, nil
}

func SetDatabaseSettings(dbType, dbDSN string) error {
	configData, err := readConfig()
	if err != nil {
		return errors.Wrap(err)
	}

	if !array.InArray(dbType, config.DBTypes) {
		return errors.New("dbtype error")
	}

	configData.DbType = dbType
	configData.DbDSN = dbDSN

	// 检查数据库是否能正确连接 todo

	err = writeConfig(configData)
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func writeConfig(cfg *config.Config) error {
	bytes, _ := json.Marshal(cfg)
	err := os.WriteFile("./config/config.json", bytes, 0666)
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func readConfig() (*config.Config, error) {
	configData := config.Config{
		DkimPrivateKeyPath: "config/dkim/dkim.priv",
		SSLPrivateKeyPath:  "config/ssl/private.key",
		SSLPublicKeyPath:   "config/ssl/public.crt",
	}
	if !file.PathExist("./config/config.json") {
		bytes, _ := json.Marshal(configData)
		err := os.WriteFile("./config/config.json", bytes, 0666)
		if err != nil {
			return nil, errors.Wrap(err)
		}
	} else {
		cfgData, err := os.ReadFile("./config/config.json")
		if err != nil {
			return nil, errors.Wrap(err)
		}

		err = json.Unmarshal(cfgData, &configData)
		if err != nil {
			return nil, errors.Wrap(err)
		}
	}
	return &configData, nil
}
