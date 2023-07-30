package main

import (
	"bytes"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"pmail/config"
	"pmail/dto"
	"pmail/dto/parsemail"
	"pmail/hooks"
	"pmail/http_server"
	"pmail/mysql"
	"pmail/session"
	"pmail/smtp_server"
	"time"
)

type logFormatter struct {
}

// Format 定义日志输出格式
func (l *logFormatter) Format(entry *log.Entry) ([]byte, error) {
	b := bytes.Buffer{}

	b.WriteString(fmt.Sprintf("[%s]", entry.Level.String()))
	b.WriteString(fmt.Sprintf("[%s]", entry.Time.Format("2006-01-02 15:04:05")))
	if entry.Context != nil {
		b.WriteString(fmt.Sprintf("[%s]", entry.Context.(*dto.Context).GetValue(dto.LogID)))
	}
	b.WriteString(fmt.Sprintf("[%s:%d]", entry.Caller.File, entry.Caller.Line))
	b.WriteString(entry.Message)

	b.WriteString("\n")
	return b.Bytes(), nil
}

var (
	gitHash   string
	buildTime string
	goVersion string
)

func main() {
	// 设置日志格式为json格式
	//log.SetFormatter(&log.JSONFormatter{})

	log.SetFormatter(&logFormatter{})
	log.SetReportCaller(true)

	// 设置将日志输出到标准输出（默认的输出为stderr,标准错误）
	// 日志消息输出可以是任意的io.writer类型
	log.SetOutput(os.Stdout)

	// 设置日志级别为warn以上
	log.SetLevel(log.DebugLevel)
	var cst, _ = time.LoadLocation("Asia/Shanghai")
	time.Local = cst

	config.Init()
	parsemail.Init()
	mysql.Init()
	session.Init()
	hooks.Init()

	// smtp server start
	go smtp_server.Start()

	// http server start
	go http_server.Start()

	log.Infoln("***************************************************")
	log.Infoln("***\tServer Start Success Version:1.0.0")
	log.Infof("***\tGit Commit Hash: %s ", gitHash)
	log.Infof("***\tBuild TimeStamp: %s ", buildTime)
	log.Infof("***\tBuild GoLang Version: %s ", goVersion)
	log.Infoln("***************************************************")

	s := make(chan bool)
	<-s
}
