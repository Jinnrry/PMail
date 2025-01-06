package ssl

import (
	"fmt"
	"github.com/Jinnrry/pmail/config"
	"testing"
)

func TestCheckSSLCrtInfo(t *testing.T) {
	config.Init()

	got, got1, match, err := CheckSSLCrtInfo()

	fmt.Println(got, got1, match, err)
}
