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
	"io"
	"io/fs"
	"os"
	"strings"

	"github.com/Jinnrry/pmail/config"
	"github.com/Jinnrry/pmail/utils/consts"
	"github.com/Jinnrry/pmail/utils/context"
	"github.com/emersion/go-msgauth/dkim"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ed25519"
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

// 返回 path 处私钥对应的 DKIM TXT 记录
func DkimRecord(path string) (string, error) {
	key, err := loadPrivateKey(path)
	if errors.Is(err, fs.ErrNotExist) {
		// Go 1.24 起此操作不会失败
		rsaKey, _ := rsa.GenerateKey(rand.Reader, 2048)
		der, _ := x509.MarshalPKCS8PrivateKey(rsaKey)
		err = os.WriteFile(path, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}), 0600)
		key = rsaKey
	}
	if err != nil {
		return "", err
	}

	switch pub := key.Public().(type) {
	case *rsa.PublicKey:
		// RFC 6376 勘误 3017 允许 PKIX 和 PKCS1 两种格式
		// 与 opendkim、Gmail 一致用 PKIX
		der, _ := x509.MarshalPKIXPublicKey(pub)
		return "v=DKIM1; k=rsa; p=" + base64.StdEncoding.EncodeToString(der), nil
	case ed25519.PublicKey:
		return "v=DKIM1; k=ed25519; p=" + base64.StdEncoding.EncodeToString(pub), nil
	}
	panic("unreachable")
}

// 加载 path 处的 DKIM 签名私钥
func loadPrivateKey(path string) (crypto.Signer, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(b)
	if block == nil {
		return nil, fmt.Errorf("no PEM data found")
	}

	var key any
	switch strings.ToUpper(block.Type) {
	case "PRIVATE KEY":
		key, err = x509.ParsePKCS8PrivateKey(block.Bytes)
	case "RSA PRIVATE KEY":
		key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	case "EDDSA PRIVATE KEY":
		if len(block.Bytes) != ed25519.PrivateKeySize {
			return nil, fmt.Errorf("invalid Ed25519 private key size")
		}
		key = ed25519.PrivateKey(block.Bytes)
	default:
		return nil, fmt.Errorf("unknown private key type: '%v'", block.Type)
	}
	if err != nil {
		return nil, err
	}

	switch k := key.(type) {
	case *rsa.PrivateKey:
		if k.N.BitLen() < 1024 {
			return nil, fmt.Errorf("RSA key too short: %d bits", k.N.BitLen())
		}
		return k, nil
	case ed25519.PrivateKey:
		return k, nil
	}
	return nil, fmt.Errorf("unsupported DKIM key type %T", key)
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
