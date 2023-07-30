package config

import (
	"embed"
	"encoding/json"
	"io/fs"
	"os"
	"strings"
)

type Config struct {
	Domain             string `json:"domain"`
	DkimPrivateKeyPath string `json:"dkimPrivateKeyPath"`
	SSLPrivateKeyPath  string `json:"SSLPrivateKeyPath"`
	SSLPublicKeyPath   string `json:"SSLPublicKeyPath"`
	MysqlDSN           string `json:"mysqlDSN"`

	WeChatPushAppId      string `json:"weChatPushAppId"`
	WeChatPushSecret     string `json:"weChatPushSecret"`
	WeChatPushTemplateId string `json:"weChatPushTemplateId"`
	WeChatPushUserId     string `json:"weChatPushUserId"`

	Tables         map[string]string
	TablesInitData map[string]string
}

//go:embed tables/*
var tableConfig embed.FS

var Instance *Config

func Init() {
	var cfgData []byte
	var err error
	args := os.Args

	if len(args) >= 2 && args[len(args)-1] == "dev" {
		cfgData, err = os.ReadFile("./config/config.dev.json")
		if err != nil {
			panic("dev环境配置文件加载失败" + err.Error())
		}
	} else {
		cfgData, err = os.ReadFile("./config/config.json")
		if err != nil {
			panic("配置文件加载失败" + err.Error())
		}
	}

	err = json.Unmarshal(cfgData, &Instance)
	if err != nil {
		panic("配置文件加载失败" + err.Error())
	}

	// 读取表设置
	Instance.Tables = map[string]string{}
	Instance.TablesInitData = map[string]string{}

	err = fs.WalkDir(tableConfig, "tables", func(path string, info fs.DirEntry, err error) error {
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".sql") {
			tableName := strings.ReplaceAll(info.Name(), ".sql", "")
			i, e := tableConfig.ReadFile(path)
			if e != nil {
				panic(e)
			}
			if strings.Contains(path, "data") {
				Instance.TablesInitData[tableName] = string(i)
			} else {
				Instance.Tables[tableName] = string(i)
			}

		}
		return nil
	})

	if err != nil {
		panic(err)
	}

}
