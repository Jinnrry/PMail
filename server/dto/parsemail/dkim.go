package parsemail

import (
	"bytes"
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/emersion/go-msgauth/dkim"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ed25519"
	"io"
	"os"
	"pmail/config"
	"strings"
)

type Dkim struct {
	privateKey crypto.Signer
}

var instance *Dkim

func Init() {
	privateKey, err := loadPrivateKey(config.Instance.DkimPrivateKeyPath)
	if err != nil {
		panic("DKIM load fail! Please set dkim!  dkim私钥加载失败！请先设置dkim秘钥")
	}

	instance = &Dkim{
		privateKey: privateKey,
	}
}

func loadPrivateKey(path string) (crypto.Signer, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(b)
	if block == nil {
		return nil, fmt.Errorf("no PEM data found")
	}

	switch strings.ToUpper(block.Type) {
	case "PRIVATE KEY":
		k, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		return k.(crypto.Signer), nil
	case "RSA PRIVATE KEY":
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	case "EDDSA PRIVATE KEY":
		if len(block.Bytes) != ed25519.PrivateKeySize {
			return nil, fmt.Errorf("invalid Ed25519 private key size")
		}
		return ed25519.PrivateKey(block.Bytes), nil
	default:
		return nil, fmt.Errorf("unknown private key type: '%v'", block.Type)
	}
}

func (p *Dkim) Sign(msgData string) []byte {
	var b bytes.Buffer
	r := strings.NewReader(msgData)

	options := &dkim.SignOptions{
		Domain:   config.Instance.Domain,
		Selector: "default",
		Signer:   p.privateKey,
	}

	if err := dkim.Sign(&b, r, options); err != nil {
		log.Fatal(err)
	}
	return b.Bytes()
}

func Check(mail io.Reader) bool {

	verifications, err := dkim.Verify(mail)
	if err != nil {
		log.Println(err)
	}

	for _, v := range verifications {
		if v.Err == nil {
			log.Println("Valid signature for:", v.Domain)
		} else {
			log.Println("Invalid signature for:", v.Domain, v.Err)
			return false
		}
	}
	return true
}
