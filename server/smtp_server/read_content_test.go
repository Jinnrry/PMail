package smtp_server

import (
	"bytes"
	log "github.com/sirupsen/logrus"
	"io/fs"
	"net"
	"net/netip"
	"os"
	"path/filepath"
	"pmail/config"
	"pmail/db"
	parsemail2 "pmail/dto/parsemail"
	"pmail/session"
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

}

func TestSession_Data(t *testing.T) {
	testInit()
	s := Session{
		RemoteAddress: net.TCPAddrFromAddrPort(netip.AddrPortFrom(netip.AddrFrom4([4]byte{}), 25)),
	}

	filepath.WalkDir("docs", func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			data, _ := os.ReadFile(path)
			s.Data(bytes.NewReader(data))
		}
		return nil
	})

}

func TestSession_DataGmail(t *testing.T) {
	testInit()
	s := Session{
		RemoteAddress: net.TCPAddrFromAddrPort(netip.AddrPortFrom(netip.AddrFrom4([4]byte{}), 25)),
	}

	data, _ := os.ReadFile("docs/gmail/带附件带图片.txt")
	s.Data(bytes.NewReader(data))

}
