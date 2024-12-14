package ssl

import (
	"fmt"
	"github.com/Jinnrry/pmail/config"
	"testing"
)

func TestCheckSSLCrtInfo(t *testing.T) {
	config.Init()

	got, got1, _, err := CheckSSLCrtInfo()

	fmt.Println(got, got1, err)
}
