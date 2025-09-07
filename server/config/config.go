package config

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"github.com/Jinnrry/pmail/utils/context"
	"github.com/Jinnrry/pmail/utils/errors"
	"github.com/Jinnrry/pmail/utils/file"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

var IsInit bool

type Config struct {
	LogLevel             string            `json:"logLevel"` // 日志级别
	Domain               string            `json:"domain"`
	Domains              []string          `json:"domains"` //多域名设置，把所有收信域名都填进去
	WebDomain            string            `json:"webDomain"`
	DkimPrivateKeyPath   string            `json:"dkimPrivateKeyPath"`
	SSLType              string            `json:"sslType"` // 0表示自动生成证书，HTTP挑战模式，1表示用户上传证书，2表示自动-DNS挑战模式
	SSLPrivateKeyPath    string            `json:"SSLPrivateKeyPath"`
	SSLPublicKeyPath     string            `json:"SSLPublicKeyPath"`
	DbDSN                string            `json:"dbDSN"`
	DbType               string            `json:"dbType"`
	HttpsEnabled         int               `json:"httpsEnabled"`    //后台页面是否启用https，0默认（启用），1启用，2不启用
	SpamFilterLevel      int               `json:"spamFilterLevel"` //垃圾邮件过滤级别，0不过滤、1 spf dkim 校验均失败时过滤，2 spf校验不通过时过滤
	HttpPort             int               `json:"httpPort"`        //http服务端口设置，默认80
	HttpsPort            int               `json:"httpsort"`        //https服务端口，默认443
	FrontendPort         int               `json:"frontendPort"`    //前端服务端口
	SMTPPort             int               `json:"smtpPort"`
	IMAPPort             int               `json:"imapPort"`
	POP3Port             int               `json:"pop3Port"`
	SMTPSPort            int               `json:"smtpsPort"`
	IMAPSPort            int               `json:"imapsPort"`
	POP3SPort            int               `json:"pop3sPort"`
	WeChatPushAppId      string            `json:"weChatPushAppId"`
	WeChatPushSecret     string            `json:"weChatPushSecret"`
	WeChatPushTemplateId string            `json:"weChatPushTemplateId"`
	WeChatPushUserId     string            `json:"weChatPushUserId"`
	TgBotToken           string            `json:"tgBotToken"`
	TgChatId             string            `json:"tgChatId"`
	IsInit               bool              `json:"isInit"`
	WebPushUrl           string            `json:"webPushUrl"`
	WebPushToken         string            `json:"webPushToken"`
	Tables               map[string]string `json:"-"`
	TablesInitData       map[string]string `json:"-"`
	setupPort            int               // 初始化阶段端口
}

func (c *Config) GetSetupPort() int {
	return c.setupPort
}

func (c *Config) SetSetupPort(setupPort int) {
	c.setupPort = setupPort
}

const DBTypeMySQL = "mysql"
const DBTypeSQLite = "sqlite"
const DBTypePostgres = "postgres"
const SSLTypeAutoHTTP = "0" //自动生成证书
const SSLTypeAutoDNS = "2"  //自动生成证书，DNS api验证
const SSLTypeUser = "1"     //用户上传证书

var DBTypes []string = []string{DBTypeMySQL, DBTypeSQLite, DBTypePostgres}

var Instance *Config = &Config{}

type logFormatter struct {
}

// Format 定义日志输出格式
func (l *logFormatter) Format(entry *log.Entry) ([]byte, error) {
	b := bytes.Buffer{}

	b.WriteString(fmt.Sprintf("[%s]", entry.Level.String()))
	b.WriteString(fmt.Sprintf("[%s]", entry.Time.Format("2006-01-02 15:04:05")))
	if entry.Context != nil {
		ctx := entry.Context.(*context.Context)
		if ctx != nil {
			b.WriteString(fmt.Sprintf("[%s]", ctx.GetValue(context.LogID)))
		}
	}
	b.WriteString(fmt.Sprintf("[%s:%d]", entry.Caller.File, entry.Caller.Line))
	b.WriteString(entry.Message)

	b.WriteString("\n")
	return b.Bytes(), nil
}
func Init() {
	wd, _ := os.Getwd()
	logrus.Infof("start init config, work dir: %s", wd)
	// 检查环境变量中是否有指定配置文件
	configPath := os.Getenv("PMAIL_CONFIG_PATH")
	if configPath == "" {
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
			log.Errorf("config file not found,%s", err.Error())
			return
		}
	}

	err = json.Unmarshal(cfgData, &Instance)
	if err != nil {
		return
	}

	if len(Instance.Domains) == 0 && Instance.Domain != "" {
		Instance.Domains = []string{Instance.Domain}
	}

	if Instance.Domain != "" && Instance.IsInit {
		IsInit = true
	}

	// 设置日志格式为json格式
	log.SetFormatter(&logFormatter{})
	log.SetReportCaller(true)

	// 设置将日志输出到标准输出（默认的输出为stderr,标准错误）
	// 日志消息输出可以是任意的io.writer类型
	log.SetOutput(os.Stdout)

	var cstZone = time.FixedZone("CST", 8*3600)
	time.Local = cstZone
	if Instance != nil {
		switch Instance.LogLevel {
		case "":
			log.SetLevel(log.InfoLevel)
		case "debug":
			log.SetLevel(log.DebugLevel)
		case "info":
			log.SetLevel(log.InfoLevel)
		case "warn":
			log.SetLevel(log.WarnLevel)
		case "error":
			log.SetLevel(log.ErrorLevel)
		default:
			log.SetLevel(log.InfoLevel)
		}
	} else {
		log.SetLevel(log.InfoLevel)
	}

}

func ReadPrivateKey() (*ecdsa.PrivateKey, bool) {
	key, err := os.ReadFile("./config/ssl/account_private.pem")
	if err != nil {
		return createNewPrivateKey(), true
	}

	block, _ := pem.Decode(key)
	x509Encoded := block.Bytes
	privateKey, _ := x509.ParseECPrivateKey(x509Encoded)

	return privateKey, false
}

func createNewPrivateKey() *ecdsa.PrivateKey {
	// Create a user. New accounts need an email and private key to start.
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}
	x509Encoded, _ := x509.MarshalECPrivateKey(privateKey)

	// 将ec 密钥写入到 pem文件里
	keypem, _ := os.OpenFile("./config/ssl/account_private.pem", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	err = pem.Encode(keypem, &pem.Block{Type: "EC PRIVATE KEY", Bytes: x509Encoded})
	if err != nil {
		panic(err)
	}
	return privateKey
}

func WriteConfig(cfg *Config) error {
	bytes, _ := json.Marshal(cfg)
	_ = os.MkdirAll("./config/", 0755)
	err := os.WriteFile("./config/config.json", bytes, 0666)
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func ReadConfig() (*Config, error) {
	configData := Config{
		DkimPrivateKeyPath: "config/dkim/dkim.priv",
		SSLPrivateKeyPath:  "config/ssl/private.key",
		SSLPublicKeyPath:   "config/ssl/public.crt",
		HttpPort:           80,
		HttpsPort:          443,
		FrontendPort:       5173,
		SMTPPort:           25,
		IMAPPort:           143,
		POP3Port:           110,
		SMTPSPort:          465,
		IMAPSPort:          993,
		POP3SPort:          995,
	}
	if !file.PathExist("./config/config.json") {
		bytes, _ := json.Marshal(configData)
		_ = os.MkdirAll("./config/", 0755)
		err := os.WriteFile("./config/config.json", bytes, 0666)
		if err != nil {
			log.Errorf("Write Config Error:%s", err.Error())
			return nil, errors.Wrap(err)
		}
	} else {
		cfgData, err := os.ReadFile("./config/config.json")
		if err != nil {
			log.Errorf("Read Config Error:%s", err.Error())
			return nil, errors.Wrap(err)
		}

		err = json.Unmarshal(cfgData, &configData)
		if err != nil {
			log.Errorf("Read Config Unmarshal Error:%s", err.Error())
			return nil, errors.Wrap(err)
		}
	}
	return &configData, nil
}
