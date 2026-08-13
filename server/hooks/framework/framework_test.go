package framework

import (
	"encoding/json"
	"testing"

	"github.com/Jinnrry/pmail/dto/parsemail"
)

func TestHookDTOIncludesEmailAuthentication(t *testing.T) {
	dto := HookDTO{
		Email: &parsemail.Email{
			Subject:        "认证结果测试",
			Authentication: parsemail.NewEmailAuthentication(false, false),
		},
	}

	body, err := json.Marshal(dto)
	if err != nil {
		t.Fatalf("序列化 HookDTO 失败：%v", err)
	}

	var decoded struct {
		Email *struct {
			Authentication *parsemail.EmailAuthentication
		}
	}
	if err = json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("反序列化 HookDTO 失败：%v", err)
	}
	if decoded.Email == nil {
		t.Fatal("HookDTO 中缺少邮件")
	}
	authentication := decoded.Email.Authentication
	if authentication == nil {
		t.Fatal("HookDTO 中缺少邮件认证结果")
	}
	if authentication.SPFPassed || authentication.DKIMPassed || !authentication.Dangerous {
		t.Fatalf("HookDTO 中的认证结果不正确：%+v", authentication)
	}
}

func TestOlderPluginEmailShapeIgnoresAuthentication(t *testing.T) {
	dto := HookDTO{
		Email: &parsemail.Email{
			Subject:        "兼容性测试",
			Authentication: parsemail.NewEmailAuthentication(true, false),
		},
	}
	body, err := json.Marshal(dto)
	if err != nil {
		t.Fatalf("序列化 HookDTO 失败：%v", err)
	}

	var oldDTO struct {
		Email *struct {
			Subject string
		}
	}
	if err = json.Unmarshal(body, &oldDTO); err != nil {
		t.Fatalf("旧插件结构无法读取新增字段后的 HookDTO：%v", err)
	}
	if oldDTO.Email == nil || oldDTO.Email.Subject != "兼容性测试" {
		t.Fatalf("旧插件结构读取到的邮件不正确：%+v", oldDTO.Email)
	}
}
