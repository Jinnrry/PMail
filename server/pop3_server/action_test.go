package pop3_server

import (
	"fmt"
	"github.com/Jinnrry/gopop"
	log "github.com/sirupsen/logrus"
	"os"
	"pmail/config"
	"pmail/db"
	parsemail2 "pmail/dto/parsemail"
	"pmail/hooks"
	"pmail/session"
	"pmail/utils/context"
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
	parsemail2.Init()
	db.Init()
	session.Init()
	hooks.Init("dev")
}

func Test_action_Stat(t *testing.T) {
	testInit()
	act := action{}
	v1, v2, v3 := act.Stat(&gopop.Session{
		Ctx: &context.Context{},
	})
	fmt.Println(v1, v2, v3)
}
