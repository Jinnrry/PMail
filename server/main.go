package main

import (
	"bytes"
	oc "context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"log/slog"
	"os"
	"pmail/config"
	"pmail/cron_server"
	"pmail/res_init"
	"pmail/utils/context"
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

var (
	gitHash   string
	buildTime string
	goVersion string
)

type logHandler struct {
	slog.Handler
	logLevel slog.Level
}

func newSLogger(level slog.Level) *slog.Logger {
	handler := &logHandler{
		// handler 使用 slog 默认的 Handler
		slog.Default().Handler(),
		level,
	}

	return slog.New(handler)
}

func (dh *logHandler) Enabled(ctx oc.Context, l slog.Level) bool {
	return l >= dh.logLevel
}

func main() {
	// 设置日志格式为json格式
	log.SetFormatter(&logFormatter{})
	log.SetReportCaller(true)

	// 设置将日志输出到标准输出（默认的输出为stderr,标准错误）
	// 日志消息输出可以是任意的io.writer类型
	log.SetOutput(os.Stdout)

	var cst, _ = time.LoadLocation("Asia/Shanghai")
	time.Local = cst

	config.Init()

	if config.Instance != nil {
		switch config.Instance.LogLevel {
		case "":
			log.SetLevel(log.InfoLevel)
		case "debug":
			slog.SetDefault(newSLogger(slog.LevelDebug))
			log.SetLevel(log.DebugLevel)
		case "info":
			log.SetLevel(log.InfoLevel)
		case "warn":
			slog.SetDefault(newSLogger(slog.LevelWarn))
			log.SetLevel(log.WarnLevel)
		case "error":
			slog.SetDefault(newSLogger(slog.LevelError))
			log.SetLevel(log.ErrorLevel)
		default:
			log.SetLevel(log.InfoLevel)
		}
	} else {
		log.SetLevel(log.InfoLevel)
	}

	log.Infoln("***************************************************")
	log.Infof("***\tServer Start Success Version:%s\n", config.Version)
	log.Infof("***\tGit Commit Hash: %s ", gitHash)
	log.Infof("***\tBuild TimeStamp: %s ", buildTime)
	log.Infof("***\tBuild GoLang Version: %s ", goVersion)
	log.Infoln("***************************************************")

	// 定时任务启动
	go cron_server.Start()

	// 核心服务启动
	res_init.Init()

	s := make(chan bool)
	<-s
}
