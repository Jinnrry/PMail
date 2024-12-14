package cron_server

import (
	"github.com/Jinnrry/pmail/config"
	"github.com/Jinnrry/pmail/services/setup/ssl"
	"github.com/Jinnrry/pmail/signal"
	log "github.com/sirupsen/logrus"
	"time"
)

var expiredTime time.Time

func Start() {

	// 第一次启动，等待到初始化完成
	if config.Instance == nil || config.IsInit == false {
		for {
			time.Sleep(1 * time.Minute)
			if config.Instance != nil && config.IsInit {
				break
			}
		}
	}

	if config.Instance.SSLType == config.SSLTypeAutoHTTP || config.Instance.SSLType == config.SSLTypeAutoDNS {
		go sslUpdateLoop()
	} else {
		go sslCheck()
	}

}

// 每天检查一遍SSL证书是否更新，更新就重启
func sslCheck() {
	var err error
	_, expiredTime, _, err = ssl.CheckSSLCrtInfo()
	if err != nil {
		panic(err)
	}

	for {
		time.Sleep(24 * time.Hour)
		_, newExpTime, _, err := ssl.CheckSSLCrtInfo()
		if err != nil {
			log.Errorf("SSL Check Error! %+v", err)
		}
		if newExpTime != expiredTime {
			expiredTime = newExpTime
			log.Infoln("SSL certificate had update! restarting")
			signal.RestartChan <- true
		}

	}
}

// 每天检查一遍SSL证书是否即将过期，即将过期就重新生成
func sslUpdateLoop() {
	for {
		ssl.Update(true)
		// 每24小时检测一次证书有效期
		time.Sleep(24 * time.Hour)
	}
}
