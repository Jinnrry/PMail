package cron_server

import (
	log "github.com/sirupsen/logrus"
	"pmail/config"
	"pmail/services/setup/ssl"
	"pmail/signal"
	"time"
)

func Start() {
	for {
		if config.Instance.IsInit {
			days, err := ssl.CheckSSLCrtInfo()
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
