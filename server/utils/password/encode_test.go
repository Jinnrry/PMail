package password

import (
	"fmt"
	"testing"
)

func TestEncode(t *testing.T) {
	fmt.Println(Encode("admin"))
}
