package config

import (
	"embed"
	"encoding/json"
	"io/fs"
	"os"
	"strings"
)

var IsInit bool

type Config struct {
	LogLevel             string            `json:"logLevel"` // 日志级别
	Domain               string            `json:"domain"`
	WebDomain            string            `json:"webDomain"`
	DkimPrivateKeyPath   string            `json:"dkimPrivateKeyPath"`
	SSLType              string            `json:"sslType"` // 0表示自动生成证书，1表示用户上传证书
	SSLPrivateKeyPath    string            `json:"SSLPrivateKeyPath"`
	SSLPublicKeyPath     string            `json:"SSLPublicKeyPath"`
	DbDSN                string            `json:"dbDSN"`
	DbType               string            `json:"dbType"`
	HttpsEnabled         int               `json:"httpsEnabled"`    //后台页面是否启用https，0默认（启用），1启用，2不启用
	SpamFilterLevel      int               `json:"spamFilterLevel"` //垃圾邮件过滤级别，0不过滤、1 spf dkim 校验均失败时过滤，2 spf校验不通过时过滤
	HttpPort             int               `json:"httpPort"`        //http服务端口设置，默认80
	HttpsPort            int               `json:"httpsPort"`       //https服务端口，默认443
	WeChatPushAppId      string            `json:"weChatPushAppId"`
	WeChatPushSecret     string            `json:"weChatPushSecret"`
	WeChatPushTemplateId string            `json:"weChatPushTemplateId"`
	WeChatPushUserId     string            `json:"weChatPushUserId"`
	TgBotToken           string            `json:"tgBotToken"`
	TgChatId             string            `json:"tgChatId"`
	IsInit               bool              `json:"isInit"`
	Tables               map[string]string `json:"-"`
	TablesInitData       map[string]string `json:"-"`
}

//go:embed tables/*
var tableConfig embed.FS

const Version = "2.2.7"

const DBTypeMySQL = "mysql"
const DBTypeSQLite = "sqlite"
const SSLTypeAuto = "0" //自动生成证书
const SSLTypeUser = "1" //用户上传证书

var DBTypes []string = []string{DBTypeMySQL, DBTypeSQLite}

var Instance *Config

func Init() {
	var cfgData []byte
	var err error
	args := os.Args

	if len(args) >= 2 && args[len(args)-1] == "dev" {
		cfgData, err = os.ReadFile("./config/config.dev.json")
		if err != nil {
			return
		}
	} else {
		cfgData, err = os.ReadFile("./config/config.json")
		if err != nil {
			return
		}
	}

	err = json.Unmarshal(cfgData, &Instance)
	if err != nil {
		return
	}

	// 读取表设置
	Instance.Tables = map[string]string{}
	Instance.TablesInitData = map[string]string{}

	root := "tables/mysql"
	if Instance.DbType == DBTypeSQLite {
		root = "tables/sqlite"
	}
	err = fs.WalkDir(tableConfig, root, func(path string, info fs.DirEntry, err error) error {
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

	if Instance.Domain != "" && Instance.IsInit {
		IsInit = true
	}

}
