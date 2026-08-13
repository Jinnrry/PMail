package smtp_server

import (
	"bytes"
	"net"
	"net/netip"
	"testing"

	"github.com/Jinnrry/pmail/config"
	"github.com/Jinnrry/pmail/db"
	"github.com/Jinnrry/pmail/dto/parsemail"
	"github.com/Jinnrry/pmail/hooks"
	"github.com/Jinnrry/pmail/hooks/framework"
	"github.com/Jinnrry/pmail/models"
	"github.com/Jinnrry/pmail/utils/context"
	"xorm.io/xorm"
)

type authenticationCaptureHook struct {
	parseAfter *parsemail.EmailAuthentication
	saveAfter  *parsemail.EmailAuthentication
}

func (h *authenticationCaptureHook) SendBefore(ctx *context.Context, email *parsemail.Email) {}

func (h *authenticationCaptureHook) SendAfter(ctx *context.Context, email *parsemail.Email, err map[string]error) {
}

func (h *authenticationCaptureHook) ReceiveParseBefore(ctx *context.Context, email *[]byte) {}

func (h *authenticationCaptureHook) ReceiveParseAfter(ctx *context.Context, email *parsemail.Email) {
	h.parseAfter = cloneAuthentication(email.Authentication)
	// 主程序认证结果不能被插件修改后带入数据库和后续钩子。
	email.Authentication.Dangerous = false
}

func (h *authenticationCaptureHook) ReceiveSaveAfter(ctx *context.Context, email *parsemail.Email, ue []*models.UserEmail) {
	h.saveAfter = cloneAuthentication(email.Authentication)
}

func (h *authenticationCaptureHook) GetName(ctx *context.Context) string {
	return "authentication_capture"
}

func (h *authenticationCaptureHook) SettingsHtml(ctx *context.Context, url string, requestData string) string {
	return ""
}

func cloneAuthentication(authentication *parsemail.EmailAuthentication) *parsemail.EmailAuthentication {
	if authentication == nil {
		return nil
	}
	ret := *authentication
	return &ret
}

func TestReceiveHooksAndDatabaseShareAuthenticationResult(t *testing.T) {
	engine := newAuthenticationTestEngine(t)

	oldDB := db.Instance
	oldConfig := config.Instance
	oldHooks := hooks.HookList
	t.Cleanup(func() {
		db.Instance = oldDB
		config.Instance = oldConfig
		hooks.HookList = oldHooks
	})

	db.Instance = engine
	config.Instance = &config.Config{
		Domain:          "test.domain",
		Domains:         []string{"test.domain"},
		SpamFilterLevel: 0,
	}
	recipient := &models.User{Account: "recipient", Name: "收件人"}
	if _, err := engine.Insert(recipient); err != nil {
		t.Fatalf("创建测试收件人失败：%v", err)
	}

	capture := &authenticationCaptureHook{}
	hooks.HookList = map[string]framework.EmailHook{"capture": capture}

	emailData := []byte("From: sender\r\nTo: recipient@test.domain\r\nSubject: authentication flow\r\n\r\nbody")
	session := Session{
		RemoteAddress: net.TCPAddrFromAddrPort(netip.MustParseAddrPort("203.0.113.1:25")),
		Ctx:           &context.Context{},
		To:            []string{"recipient@test.domain"},
	}
	if err := session.Data(bytes.NewReader(emailData)); err != nil {
		t.Fatalf("收信流程执行失败：%v", err)
	}

	assertDangerousAuthentication(t, "ReceiveParseAfter", capture.parseAfter)
	assertDangerousAuthentication(t, "ReceiveSaveAfter", capture.saveAfter)

	var saved models.Email
	has, err := engine.Where("subject = ?", "authentication flow").Get(&saved)
	if err != nil {
		t.Fatalf("查询落库邮件失败：%v", err)
	}
	if !has {
		t.Fatal("未找到落库邮件")
	}
	if saved.SPFCheck != 0 || saved.DKIMCheck != 0 {
		t.Fatalf("落库认证结果不正确：SPF=%d DKIM=%d", saved.SPFCheck, saved.DKIMCheck)
	}
}

func TestSaveSentEmailIgnoresReceiveAuthentication(t *testing.T) {
	engine := newAuthenticationTestEngine(t)

	oldDB := db.Instance
	t.Cleanup(func() { db.Instance = oldDB })
	db.Instance = engine

	email := &parsemail.Email{
		From:           &parsemail.User{EmailAddress: "sender@test.domain"},
		Subject:        "sent authentication isolation",
		Authentication: parsemail.NewEmailAuthentication(false, false),
	}
	_, saved, err := saveEmail(&context.Context{UserID: 1}, 0, email, 1, 1, nil, true, true)
	if err != nil {
		t.Fatalf("保存发件邮件失败：%v", err)
	}
	if saved.SPFCheck != 1 || saved.DKIMCheck != 1 {
		t.Fatalf("发件邮件认证落库值被收信认证对象覆盖：SPF=%d DKIM=%d", saved.SPFCheck, saved.DKIMCheck)
	}
}

func newAuthenticationTestEngine(t *testing.T) *xorm.Engine {
	t.Helper()
	engine, err := xorm.NewEngine("sqlite", "file:"+t.Name()+"?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("创建测试数据库失败：%v", err)
	}
	engine.SetMaxOpenConns(1)
	if err = engine.Sync2(&models.User{}, &models.Email{}, &models.UserEmail{}, &models.Rule{}); err != nil {
		t.Fatalf("初始化测试数据库失败：%v", err)
	}
	t.Cleanup(func() { _ = engine.Close() })
	return engine
}

func assertDangerousAuthentication(t *testing.T, hookName string, authentication *parsemail.EmailAuthentication) {
	t.Helper()
	if authentication == nil {
		t.Fatalf("%s 未收到认证结果", hookName)
	}
	if authentication.SPFPassed || authentication.DKIMPassed || !authentication.Dangerous {
		t.Fatalf("%s 收到的认证结果不正确：%+v", hookName, authentication)
	}
}
