package ssl

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/tls"
	"crypto/x509"
	"github.com/Jinnrry/pmail/config"
	"github.com/Jinnrry/pmail/services/setup"
	"github.com/Jinnrry/pmail/signal"
	"github.com/Jinnrry/pmail/utils/async"
	"github.com/Jinnrry/pmail/utils/errors"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge/dns01"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"os"
	"time"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
)

type MyUser struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func (u *MyUser) GetEmail() string {
	return u.Email
}
func (u MyUser) GetRegistration() *registration.Resource {
	return u.Registration
}
func (u *MyUser) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

func GetSSL() string {
	cfg, err := setup.ReadConfig()
	if err != nil {
		panic(err)
	}
	if cfg.SSLType == "" {
		return config.SSLTypeAutoHTTP
	}

	return cfg.SSLType
}

func SetSSL(sslType, priKey, crtKey string) error {
	cfg, err := setup.ReadConfig()
	if err != nil {
		panic(err)
	}
	if sslType == config.SSLTypeAutoHTTP || sslType == config.SSLTypeUser || sslType == config.SSLTypeAutoDNS {
		cfg.SSLType = sslType
	} else {
		return errors.New("SSL Type Error!")
	}

	if cfg.SSLType == config.SSLTypeUser {
		cfg.SSLPrivateKeyPath = priKey
		cfg.SSLPublicKeyPath = crtKey
		// 手动设置证书的情况下后台地址默认不启用https
		cfg.HttpsEnabled = 2
	}

	err = setup.WriteConfig(cfg)
	if err != nil {
		return errors.Wrap(err)
	}

	return nil
}

func renewCertificate(privateKey *ecdsa.PrivateKey, cfg *config.Config) error {

	myUser := MyUser{
		Email: "i@" + cfg.Domain,
		key:   privateKey,
	}

	conf := lego.NewConfig(&myUser)
	conf.UserAgent = "PMail"
	conf.Certificate.KeyType = certcrypto.RSA2048

	// A client facilitates communication with the CA server.
	client, err := lego.NewClient(conf)
	if err != nil {
		return errors.Wrap(err)
	}

	if cfg.SSLType == config.SSLTypeAutoHTTP {
		err = client.Challenge.SetHTTP01Provider(GetHttpChallengeInstance())
		if err != nil {
			return errors.Wrap(err)
		}
	} else if cfg.SSLType == config.SSLTypeAutoDNS {
		err = client.Challenge.SetDNS01Provider(GetDnsChallengeInstance(), dns01.AddDNSTimeout(60*time.Minute))
		if err != nil {
			return errors.Wrap(err)
		}

		log.Errorf("Please Set DNS Record/请将以下内容添加到DNS记录中:\n")
		for _, item := range GetDnsChallengeInstance().GetDNSSettings(nil) {
			log.Errorf("Type:%s\tHost:%s\tValue:%s\n", item.Type, item.Host, item.Value)
		}

	}

	var reg *registration.Resource

	reg, err = client.Registration.ResolveAccountByKey()
	if err != nil {
		return errors.Wrap(err)
	}

	myUser.Registration = reg

	domains := []string{cfg.WebDomain}
	for _, domain := range cfg.Domains {
		domains = append(domains, "smtp."+domain)
		domains = append(domains, "pop."+domain)
	}

	request := certificate.ObtainRequest{
		Domains: domains,
		Bundle:  true,
	}

	log.Infof("wait ssl renew")
	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile("./config/ssl/private.key", certificates.PrivateKey, 0666)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile("./config/ssl/public.crt", certificates.Certificate, 0666)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile("./config/ssl/issuerCert.crt", certificates.IssuerCertificate, 0666)
	if err != nil {
		panic(err)
	}

	return nil
}

func generateCertificate(privateKey *ecdsa.PrivateKey, cfg *config.Config, newAccount bool) error {

	myUser := MyUser{
		Email: "i@" + cfg.Domain,
		key:   privateKey,
	}

	conf := lego.NewConfig(&myUser)
	conf.UserAgent = "PMail"
	conf.Certificate.KeyType = certcrypto.RSA2048

	// A client facilitates communication with the CA server.
	client, err := lego.NewClient(conf)
	if err != nil {
		return errors.Wrap(err)
	}

	if cfg.SSLType == config.SSLTypeAutoHTTP {
		err = client.Challenge.SetHTTP01Provider(GetHttpChallengeInstance())
		if err != nil {
			return errors.Wrap(err)
		}
	} else if cfg.SSLType == config.SSLTypeAutoDNS {
		err = client.Challenge.SetDNS01Provider(GetDnsChallengeInstance(), dns01.AddDNSTimeout(60*time.Minute))
		if err != nil {
			return errors.Wrap(err)
		}
	}

	var reg *registration.Resource

	if newAccount {
		reg, err = client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
		if err != nil {
			return errors.Wrap(err)
		}
	} else {
		reg, err = client.Registration.ResolveAccountByKey()
		if err != nil {
			return errors.Wrap(err)
		}
	}

	myUser.Registration = reg

	domains := []string{cfg.WebDomain}
	for _, domain := range cfg.Domains {
		domains = append(domains, "smtp."+domain)
		domains = append(domains, "pop."+domain)
	}

	request := certificate.ObtainRequest{
		Domains: domains,
		Bundle:  true,
	}

	as := async.New(nil)

	as.Process(func(params any) {
		log.Infof("wait ssl")
		certificates, err := client.Certificate.Obtain(request)
		if err != nil {
			panic(err)
		}
		log.Infof("证书校验通过！")
		err = os.WriteFile("./config/ssl/private.key", certificates.PrivateKey, 0666)
		if err != nil {
			panic(err)
		}

		err = os.WriteFile("./config/ssl/public.crt", certificates.Certificate, 0666)
		if err != nil {
			panic(err)
		}

		err = os.WriteFile("./config/ssl/issuerCert.crt", certificates.IssuerCertificate, 0666)
		if err != nil {
			panic(err)
		}

		setup.Finish()

	}, nil)

	return nil
}

func GenSSL(update bool) error {

	cfg, err := setup.ReadConfig()
	if err != nil {
		panic(err)
	}

	if !update {
		privateFile, errpi := os.ReadFile(cfg.SSLPrivateKeyPath)
		public, errpu := os.ReadFile(cfg.SSLPublicKeyPath)
		// 当前存在证书数据，就不生成了
		if errpi == nil && errpu == nil && len(privateFile) > 0 && len(public) > 0 {
			return nil
		}
	}

	privateKey, newAccount := config.ReadPrivateKey()

	if !update {
		return generateCertificate(privateKey, cfg, newAccount)
	}

	return renewCertificate(privateKey, cfg)
}

// CheckSSLCrtInfo 返回证书过期剩余天数
func CheckSSLCrtInfo() (int, time.Time, error) {

	cfg, err := setup.ReadConfig()
	if err != nil {
		panic(err)
	}
	// load cert and key by tls.LoadX509KeyPair
	tlsCert, err := tls.LoadX509KeyPair(cfg.SSLPublicKeyPath, cfg.SSLPrivateKeyPath)
	if err != nil {
		return -1, time.Now(), errors.Wrap(err)
	}

	cert, err := x509.ParseCertificate(tlsCert.Certificate[0])

	if err != nil {
		return -1, time.Now(), errors.Wrap(err)
	}

	// 检查过期时间
	hours := cert.NotAfter.Sub(time.Now()).Hours()

	if hours <= 0 {
		return -1, time.Now(), errors.New("Certificate has expired")
	}

	return cast.ToInt(hours / 24), cert.NotAfter, nil
}

func Update(needRestart bool) {
	if config.Instance != nil && config.Instance.IsInit && (config.Instance.SSLType == config.SSLTypeAutoHTTP || config.Instance.SSLType == config.SSLTypeAutoDNS) {
		days, _, err := CheckSSLCrtInfo()
		if days < 30 || err != nil {
			if err != nil {
				log.Errorf("SSL Check Error, Update SSL Certificate. Error Info :%+v", err)
			} else {
				log.Infof("SSL certificate remaining time is only %d days, renew SSL certificate.", days)
			}
			err = GenSSL(true)
			if err != nil {
				log.Errorf("SSL Update Error! %+v", err)
			}
			if needRestart {
				// 更新完证书，重启服务
				signal.RestartChan <- true
			}
		} else {
			log.Debugf("SSL Check.")
		}
	}

}
