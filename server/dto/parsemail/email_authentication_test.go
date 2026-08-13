package parsemail

import (
	"testing"

	"github.com/Jinnrry/pmail/models"
)

func TestNewEmailAuthentication(t *testing.T) {
	tests := []struct {
		name       string
		spfPassed  bool
		dkimPassed bool
		dangerous  bool
	}{
		{name: "SPF和DKIM均通过", spfPassed: true, dkimPassed: true},
		{name: "仅SPF通过", spfPassed: true, dkimPassed: false},
		{name: "仅DKIM通过", spfPassed: false, dkimPassed: true},
		{name: "SPF和DKIM均未通过", spfPassed: false, dkimPassed: false, dangerous: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authentication := NewEmailAuthentication(tt.spfPassed, tt.dkimPassed)
			if authentication.SPFPassed != tt.spfPassed {
				t.Fatalf("SPFPassed = %v，期望 %v", authentication.SPFPassed, tt.spfPassed)
			}
			if authentication.DKIMPassed != tt.dkimPassed {
				t.Fatalf("DKIMPassed = %v，期望 %v", authentication.DKIMPassed, tt.dkimPassed)
			}
			if authentication.Dangerous != tt.dangerous {
				t.Fatalf("Dangerous = %v，期望 %v", authentication.Dangerous, tt.dangerous)
			}
		})
	}
}

func TestNewEmailFromModelRestoresAuthenticationForReceivedEmail(t *testing.T) {
	email := NewEmailFromModel(models.Email{
		Type:      0,
		SPFCheck:  1,
		DKIMCheck: 0,
	})

	if email.Authentication == nil {
		t.Fatal("收件邮件缺少认证结果")
	}
	if !email.Authentication.SPFPassed || email.Authentication.DKIMPassed || email.Authentication.Dangerous {
		t.Fatalf("认证结果不正确：%+v", email.Authentication)
	}
}

func TestNewEmailFromModelLeavesAuthenticationNilForSentEmail(t *testing.T) {
	email := NewEmailFromModel(models.Email{
		Type:      1,
		SPFCheck:  1,
		DKIMCheck: 1,
	})

	if email.Authentication != nil {
		t.Fatalf("发件邮件不应包含收信认证结果：%+v", email.Authentication)
	}
}
