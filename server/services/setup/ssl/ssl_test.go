package ssl

import (
	"fmt"
	"testing"
)

func TestGenSSL(t *testing.T) {
	err := GenSSL(false)
	fmt.Println(err)
}

func TestGetSSLCrtInfo(t *testing.T) {
	days, err := CheckSSLCrtInfo()

	fmt.Println(days, err)
}
