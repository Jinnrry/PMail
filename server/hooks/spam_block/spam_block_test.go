package main

import (
	"testing"

	"github.com/Jinnrry/pmail/utils/context"
)

func TestSettingsHtmlRejectsNonAdminSave(t *testing.T) {
	s := &SpamBlock{}
	result := s.SettingsHtml(&context.Context{}, "SpamBlock/save", `{"url":"https://model.example.com/predict"}`)
	if result != "No Access Privileges" {
		t.Fatalf("non-admin save result = %q, want access denial", result)
	}
	if s.cfg.ApiURL != "" {
		t.Fatal("non-admin save changed the configuration")
	}
}
