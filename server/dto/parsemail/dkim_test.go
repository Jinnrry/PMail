package parsemail

import (
	"strings"
	"testing"
)

func TestCheck(t *testing.T) {

	res := Check(nil, strings.NewReader(`Received: from jdl.ac.cn ([159.226.42.8])

xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
`))
	if res != false {
		t.Errorf("DKIM Error")
	}

}
