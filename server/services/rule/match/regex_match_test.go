package match

import (
	"pmail/models"
	"testing"
)

func TestRegexMatch_Match(t *testing.T) {
	r := NewRegexMatch("Subject", "\\d+")

	ret := r.Match(nil, &models.Email{
		Subject: "111",
	})

	if !ret {
		t.Errorf("失败")
	}
}
