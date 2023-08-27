package setup

import (
	"testing"
)

func TestSetAdminPassword(t *testing.T) {

	SetAdminPassword(nil, "admin", "admin")
}
