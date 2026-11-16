package group

import (
	"path/filepath"
	"testing"

	"github.com/Jinnrry/pmail/db"
	"github.com/Jinnrry/pmail/models"
	"github.com/Jinnrry/pmail/utils/context"
	_ "modernc.org/sqlite"
	"xorm.io/xorm"
)

func TestDeleteGroupOwnershipAndAtomicity(t *testing.T) {
	for _, scenario := range []string{"other user", "own subtree", "update failure"} {
		t.Run(scenario, func(t *testing.T) {
			engine, err := xorm.NewEngine("sqlite", filepath.Join(t.TempDir(), "groups.db"))
			if err != nil {
				t.Fatal(err)
			}
			engine.SetMaxOpenConns(1)
			if err := engine.Sync2(&models.Group{}, &models.UserEmail{}); err != nil {
				t.Fatal(err)
			}
			old := db.Instance
			db.Instance = engine
			t.Cleanup(func() { db.Instance = old; engine.Close() })
			parent := models.Group{UserId: 2, Name: "Parent"}
			if _, err := engine.Insert(&parent); err != nil {
				t.Fatal(err)
			}
			child := models.Group{UserId: 2, Name: "Child", ParentId: parent.ID}
			foreign := models.Group{UserId: 3, Name: "Foreign", ParentId: parent.ID}
			if _, err := engine.Insert(&child, &foreign); err != nil {
				t.Fatal(err)
			}
			mails := []models.UserEmail{
				{UserID: 2, EmailID: 1, GroupId: parent.ID},
				{UserID: 2, EmailID: 2, GroupId: child.ID},
				{UserID: 3, EmailID: 3, GroupId: parent.ID},
			}
			for i := range mails {
				if _, err := engine.Insert(&mails[i]); err != nil {
					t.Fatal(err)
				}
			}
			userID := 2
			if scenario == "other user" {
				userID = 1
			}
			if scenario == "update failure" {
				if _, err := engine.Exec(`CREATE TRIGGER reject_mail_update BEFORE UPDATE ON user_email BEGIN SELECT RAISE(ABORT, 'test failure'); END`); err != nil {
					t.Fatal(err)
				}
			}
			deleted, err := DelGroup(&context.Context{UserID: userID}, parent.ID)
			if scenario == "update failure" {
				if err == nil {
					t.Fatal("expected update failure")
				}
			} else if err != nil {
				t.Fatal(err)
			}
			if deleted != (scenario == "own subtree") {
				t.Fatalf("deleted=%v", deleted)
			}
			for _, original := range mails {
				var actual models.UserEmail
				if _, err := engine.ID(original.ID).Get(&actual); err != nil {
					t.Fatal(err)
				}
				want := original.GroupId
				if scenario == "own subtree" && original.UserID == 2 {
					want = 0
				}
				if actual.GroupId != want {
					t.Fatalf("mail %d group=%d, want %d", original.ID, actual.GroupId, want)
				}
			}
			for _, original := range []models.Group{parent, child, foreign} {
				found, err := engine.ID(original.ID).Exist(&models.Group{})
				if err != nil {
					t.Fatal(err)
				}
				want := scenario != "own subtree" || original.UserId == 3
				if found != want {
					t.Fatalf("group %d exists=%v, want %v", original.ID, found, want)
				}
			}
		})
	}
}
