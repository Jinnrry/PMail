package password

import (
	"crypto/md5"
	"encoding/hex"
)

// Encode 对密码两次md5加盐
func Encode(password string) string {
	encodePwd := md5Encode(md5Encode(password+"pmail") + "pmail2023")
	return encodePwd
}

func md5Encode(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}
