package parsemail

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestDecodeEmailContentFromTxt(t *testing.T) {

	c, _ := os.ReadFile("../../docs/gmail/带附件带图片.txt")

	r := strings.NewReader(string(c))

	email := NewEmailFromReader(nil, r)

	fmt.Println(email)
}

func TestDecodeEmailContentFromTxt3(t *testing.T) {

	c, _ := os.ReadFile("../../docs/pmail/带附件.txt")

	r := strings.NewReader(string(c))

	email := NewEmailFromReader(nil, r)

	fmt.Println(email)
}

func TestDecodeEmailContentFromTxt2(t *testing.T) {
	c, _ := os.ReadFile("../../docs/qqemail/带图片格式排版.txt")

	r := strings.NewReader(string(c))

	email := NewEmailFromReader(nil, r)

	fmt.Println(email)

}
