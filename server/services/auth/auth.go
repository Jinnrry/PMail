package auth

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	log "github.com/sirupsen/logrus"
	"os"
	"pmail/db"
	"pmail/models"
	"pmail/utils/context"
	"strings"
)

// HasAuth 检查当前用户是否有某个邮件的auth
func HasAuth(ctx *context.Context, email *models.Email) bool {
	// 获取当前用户的auth
	var auth []models.UserAuth
	err := db.Instance.Select(&auth, db.WithContext(ctx, "select * from user_auth where user_id = ?"), ctx.UserID)
	if err != nil {
		log.WithContext(ctx).Errorf("SQL error:%+v", err)
		return false
	}

	var hasAuth bool
	for _, userAuth := range auth {
		if userAuth.EmailAccount == "*" {
			hasAuth = true
			break
		} else if strings.Contains(email.Bcc, ctx.UserAccount) || strings.Contains(email.Cc, ctx.UserAccount) || strings.Contains(email.To, ctx.UserAccount) {
			hasAuth = true
			break
		}
	}

	return hasAuth
}

func DkimGen() string {
	privKeyStr, _ := os.ReadFile("./config/dkim/dkim.priv")
	publicKeyStr, _ := os.ReadFile("./config/dkim/dkim.public")
	if len(privKeyStr) > 0 && len(publicKeyStr) > 0 {
		return string(publicKeyStr)
	}

	var (
		privKey crypto.Signer
		err     error
	)

	privKey, err = rsa.GenerateKey(rand.Reader, 1024)

	if err != nil {
		log.Fatalf("Failed to generate key: %v", err)
	}

	privBytes, err := x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		log.Fatalf("Failed to marshal private key: %v", err)
	}

	f, err := os.OpenFile("./config/dkim/dkim.priv", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Failed to create key file: %v", err)
	}
	defer f.Close()

	privBlock := pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privBytes,
	}
	if err := pem.Encode(f, &privBlock); err != nil {
		log.Fatalf("Failed to write key PEM block: %v", err)
	}
	if err := f.Close(); err != nil {
		log.Fatalf("Failed to close key file: %v", err)
	}

	var pubBytes []byte

	switch pubKey := privKey.Public().(type) {
	case *rsa.PublicKey:
		// RFC 6376 is inconsistent about whether RSA public keys should
		// be formatted as RSAPublicKey or SubjectPublicKeyInfo.
		// Erratum 3017 (https://www.rfc-editor.org/errata/eid3017)
		// proposes allowing both.  We use SubjectPublicKeyInfo for
		// consistency with other implementations including opendkim,
		// Gmail, and Fastmail.
		pubBytes, err = x509.MarshalPKIXPublicKey(pubKey)
		if err != nil {
			log.Fatalf("Failed to marshal public key: %v", err)
		}
	default:
		panic("unreachable")
	}

	params := []string{
		"v=DKIM1",
		"k=rsa",
		"p=" + base64.StdEncoding.EncodeToString(pubBytes),
	}

	publicKey := strings.Join(params, "; ")

	os.WriteFile("./config/dkim/dkim.public", []byte(publicKey), 0666)

	return publicKey
}
