package rule

import (
	"strings"
	"testing"

	"github.com/Jinnrry/pmail/config"
	"github.com/Jinnrry/pmail/db"
	"github.com/Jinnrry/pmail/dto/parsemail"
	"github.com/Jinnrry/pmail/models"
	"github.com/Jinnrry/pmail/utils/context"
	_ "modernc.org/sqlite"
	"xorm.io/xorm"
)

func TestForwardToLocalUser(t *testing.T) {
	oldConfig := config.Instance
	oldDB := db.Instance
	defer func() {
		config.Instance = oldConfig
		db.Instance = oldDB
	}()

	config.Instance = &config.Config{
		Domain:  "example.com",
		Domains: []string{"example.com"},
	}

	engine, err := xorm.NewEngine("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()
	db.Instance = engine

	if err := engine.Sync2(&models.User{}, &models.UserEmail{}); err != nil {
		t.Fatal(err)
	}

	sender := &models.User{ID: 1, Account: "sender", Name: "sender"}
	recipient := &models.User{ID: 2, Account: "recipient", Name: "recipient", IsAdmin: 1}
	if _, err := engine.Insert(sender, recipient); err != nil {
		t.Fatal(err)
	}

	email := &parsemail.Email{MessageId: 123, Status: 0}
	if err := doForward(&context.Context{}, email, "recipient@example.com", sender, nil); err != nil {
		t.Fatal(err)
	}

	var links []models.UserEmail
	if err := engine.Table(&models.UserEmail{}).Where("email_id=?", email.MessageId).Find(&links); err != nil {
		t.Fatal(err)
	}
	if len(links) != 1 || links[0].UserID != recipient.ID {
		t.Fatalf("expected one user_email link for recipient, got %+v", links)
	}

	if err := doForward(&context.Context{}, email, "recipient@example.com", sender, nil); err != nil {
		t.Fatal(err)
	}
	var count int64
	count, err = engine.Table(&models.UserEmail{}).Where("email_id=? and user_id=?", email.MessageId, recipient.ID).Count()
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("expected idempotent local forward link, got %d links", count)
	}
}

func TestForwardToLocalUserRejectsSelfLoop(t *testing.T) {
	oldConfig := config.Instance
	defer func() { config.Instance = oldConfig }()

	config.Instance = &config.Config{
		Domain:  "example.com",
		Domains: []string{"example.com"},
	}

	err := doForward(&context.Context{}, &parsemail.Email{MessageId: 1}, "sender@example.com", &models.User{Account: "sender"}, nil)
	if err == nil || !strings.Contains(err.Error(), "loop forwarding to self") {
		t.Fatalf("expected self-loop error, got %v", err)
	}
}
