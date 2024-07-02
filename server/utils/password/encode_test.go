package password

import (
	"fmt"
	"testing"
)

func TestEncode(t *testing.T) {
	fmt.Println(Encode("user2"))
	fmt.Println(Encode("user2New"))
}
