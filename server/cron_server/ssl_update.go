package cron_server

import (
	log "github.com/sirupsen/logrus"
	"pmail/config"
	"pmail/services/setup/ssl"
	"pmail/signal"
	"time"
)

var expiredTime time.Time

func Start() {
	if config.Instance.SSLType == "0" {
		go sslUpdate()
	} else {
		go sslCheck()
	}

}

// 每天检查一遍SSL证书是否更新，更新就重启
func sslCheck() {
	var err error
	_, expiredTime, err = ssl.CheckSSLCrtInfo()
	if err != nil {
		panic(err)
	}

	for {
		time.Sleep(24 * time.Hour)
		_, newExpTime, err := ssl.CheckSSLCrtInfo()
		if err != nil {
			log.Errorf("SSL Check Error! %+v", err)
		}
		if newExpTime != expiredTime {
			log.Infoln("SSL certificate had update! restarting")
			signal.RestartChan <- true
		}

	}
}

// 每天检查一遍SSL证书是否即将过期，即将过期就重新生成
func sslUpdate() {
	for {
		if config.Instance != nil && config.Instance.IsInit && config.Instance.SSLType == "0" {
			days, _, err := ssl.CheckSSLCrtInfo()
			if days < 30 || err != nil {
				if err != nil {
					log.Errorf("SSL Check Error, Update SSL Certificate. Error Info :%+v", err)
				} else {
					log.Infof("SSL certificate remaining time is only %d days, renew SSL certificate.", days)
				}
				err = ssl.GenSSL(true)
				if err != nil {
					log.Errorf("SSL Update Error! %+v", err)
				}
				// 更新完证书，重启服务
				signal.RestartChan <- true
			} else {
				log.Debugf("SSL Check.")
			}
		}
		// 每24小时检测一次证书有效期
		time.Sleep(24 * time.Hour)
	}
}
