package setup

import (
	"encoding/json"
	"os"
	"pmail/config"
	"pmail/db"
	"pmail/models"
	"pmail/utils/array"
	"pmail/utils/context"
	"pmail/utils/errors"
	"pmail/utils/file"
	"pmail/utils/password"
)

func GetDatabaseSettings(ctx *context.Context) (string, string, error) {
	configData, err := ReadConfig()
	if err != nil {
		return "", "", errors.Wrap(err)
	}

	if configData.DbType == "" && configData.DbDSN == "" {
		return config.DBTypeSQLite, "./config/pmail.db", nil
	}

	return configData.DbType, configData.DbDSN, nil
}

func GetAdminPassword(ctx *context.Context) (string, error) {

	users := []*models.User{}
	err := db.Instance.Find(&users)
	if err != nil {
		return "", errors.Wrap(err)
	}

	if len(users) > 0 {
		return users[0].Account, nil
	}

	return "", nil
}

func SetAdminPassword(ctx *context.Context, account, pwd string) error {
	encodePwd := password.Encode(pwd)
	var user models.User = models.User{
		Account:  account,
		Name:     "admin",
		Password: encodePwd,
		IsAdmin:  1,
	}

	_, err := db.Instance.Insert(&user)
	if err != nil {
		return errors.Wrap(err)
	}

	return nil
}

func SetDatabaseSettings(ctx *context.Context, dbType, dbDSN string) error {
	configData, err := ReadConfig()
	if err != nil {
		return errors.Wrap(err)
	}

	if !array.InArray(dbType, config.DBTypes) {
		return errors.New("dbtype error")
	}

	if dbDSN == "" {
		return errors.New("DSN error")
	}

	configData.DbType = dbType
	configData.DbDSN = dbDSN

	err = WriteConfig(configData)
	if err != nil {
		return errors.Wrap(err)
	}
	config.Init()
	// 检查数据库是否能正确连接
	err = db.Init("")
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func WriteConfig(cfg *config.Config) error {
	bytes, _ := json.Marshal(cfg)
	err := os.WriteFile("./config/config.json", bytes, 0666)
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func ReadConfig() (*config.Config, error) {
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
