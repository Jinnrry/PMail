package ssl

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/spf13/cast"
	"os"
	"pmail/config"
	"pmail/services/setup"
	"pmail/utils/errors"
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
		return config.SSLTypeAuto
	}

	return cfg.SSLType
}

func SetSSL(sslType string) error {
	cfg, err := setup.ReadConfig()
	if err != nil {
		panic(err)
	}
	if sslType == config.SSLTypeAuto || sslType == config.SSLTypeUser {
		cfg.SSLType = sslType
	} else {
		return errors.New("SSL Type Error!")
	}

	err = setup.WriteConfig(cfg)
	if err != nil {
		return errors.Wrap(err)
	}

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

	// Create a user. New accounts need an email and private key to start.
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return errors.Wrap(err)
	}

	myUser := MyUser{
		Email: "i@" + cfg.Domain,
		key:   privateKey,
	}

	config := lego.NewConfig(&myUser)

	config.Certificate.KeyType = certcrypto.RSA2048

	// A client facilitates communication with the CA server.
	client, err := lego.NewClient(config)
	if err != nil {
		return errors.Wrap(err)
	}

	err = client.Challenge.SetHTTP01Provider(GetHttpChallengeInstance())
	if err != nil {
		return errors.Wrap(err)
	}

	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return errors.Wrap(err)
	}
	myUser.Registration = reg

	request := certificate.ObtainRequest{
		Domains: []string{"smtp." + cfg.Domain, cfg.WebDomain},
		Bundle:  true,
	}
	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		return errors.Wrap(err)
	}

	err = os.WriteFile("./config/ssl/private.key", certificates.PrivateKey, 0666)
	if err != nil {
		return errors.Wrap(err)
	}

	err = os.WriteFile("./config/ssl/public.crt", certificates.Certificate, 0666)
	if err != nil {
		return errors.Wrap(err)
	}

	err = os.WriteFile("./config/ssl/issuerCert.crt", certificates.IssuerCertificate, 0666)
	if err != nil {
		return errors.Wrap(err)
	}

	return nil
}

// CheckSSLCrtInfo 返回证书过期剩余天数
func CheckSSLCrtInfo() (int, error) {

	cfg, err := setup.ReadConfig()
	if err != nil {
		panic(err)
	}
	// load cert and key by tls.LoadX509KeyPair
	tlsCert, err := tls.LoadX509KeyPair(cfg.SSLPublicKeyPath, cfg.SSLPrivateKeyPath)
	if err != nil {
		return -1, errors.Wrap(err)
	}

	cert, err := x509.ParseCertificate(tlsCert.Certificate[0])

	if err != nil {
		return -1, errors.Wrap(err)
	}

	// 检查过期时间
	hours := cert.NotAfter.Sub(time.Now()).Hours()

	if hours <= 0 {
		return -1, errors.New("Certificate has expired")
	}

	return cast.ToInt(hours / 24), nil
}
