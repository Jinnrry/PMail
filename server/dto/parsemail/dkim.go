package parsemail

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/Jinnrry/pmail/config"
	"github.com/Jinnrry/pmail/utils/consts"
	"github.com/Jinnrry/pmail/utils/context"
	"github.com/emersion/go-msgauth/dkim"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ed25519"
	"io"
	"io/fs"
	"os"
	"strings"
)

type Dkim struct {
	privateKey crypto.Signer
}

var instance *Dkim

func Init() {
	privateKey, err := loadPrivateKey(config.Instance.DkimPrivateKeyPath)
	if err != nil {
		panic(config.Instance.DkimPrivateKeyPath +
			" DKIM load fail! Please set dkim!  dkim私钥加载失败！请先设置dkim秘钥" +
			err.Error())
	}

	instance = &Dkim{
		privateKey: privateKey,
	}
}

// DkimRecord returns the DKIM TXT record for the private key at path,
// generating an RSA-2048 key first if the file does not exist.
func DkimRecord(path string) (string, error) {
	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return "", err
		}
		der, _ := x509.MarshalPKCS8PrivateKey(key)
		data := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
		if err := os.WriteFile(path, data, 0600); err != nil {
			return "", err
		}
	}

	key, err := loadPrivateKey(path)
	if err != nil {
		return "", err
	}
	switch pub := key.Public().(type) {
	case *rsa.PublicKey:
		der, _ := x509.MarshalPKIXPublicKey(pub)
		return "v=DKIM1; k=rsa; p=" + base64.StdEncoding.EncodeToString(der), nil
	case ed25519.PublicKey:
		return "v=DKIM1; k=ed25519; p=" + base64.StdEncoding.EncodeToString(pub), nil
	}
	panic("unreachable")
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
		log.Errorf("%+v", err)
		return []byte(msgData)
	}
	return b.Bytes()
}

func Check(ctx *context.Context, mail io.Reader) bool {

	verifications, err := dkim.Verify(mail)
	if err != nil {
		log.WithContext(ctx).Warnf("DKIM Error:%v", err)
	}

	if len(verifications) == 0 {
		return false
	}

	for _, v := range verifications {
		if v.Domain == consts.TEST_DOMAIN {
			return true
		}
		if v.Err == nil {
			log.Println("Valid signature for:", v.Domain)
		} else {
			log.Println("Invalid signature for:", v.Domain, v.Err)
			return false
		}
	}
	return true
}
