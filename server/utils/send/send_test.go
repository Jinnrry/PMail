package send

import (
	log "github.com/sirupsen/logrus"
	"os"
	"pmail/config"
	"pmail/dto/parsemail"
	"testing"
	"time"
)

func testInit() {
	// 设置日志格式为json格式
	//log.SetFormatter(&log.JSONFormatter{})

	log.SetReportCaller(true)
	log.SetFormatter(&log.TextFormatter{
		//以下设置只是为了使输出更美观
		DisableColors:   true,
		TimestampFormat: "2006-01-02 15:03:04",
	})

	// 设置将日志输出到标准输出（默认的输出为stderr,标准错误）
	// 日志消息输出可以是任意的io.writer类型
	log.SetOutput(os.Stdout)

	// 设置日志级别为warn以上
	log.SetLevel(log.TraceLevel)

	var cst, _ = time.LoadLocation("Asia/Shanghai")
	time.Local = cst

	config.Init()
	parsemail.Init()
}
func TestSend(t *testing.T) {
	testInit()
	e := &parsemail.Email{
		From: &parsemail.User{
			Name:         "发送人",
			EmailAddress: "j@jinnrry.com",
		},
		To: []*parsemail.User{
			{"ok@jinnrry.com", "名"},
		},
		Subject: "插件测试",
		Text:    []byte("这是Text"),
		HTML:    []byte("<div>这是Html</div>"),
	}
	Send(nil, e)
}
