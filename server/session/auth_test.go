package session

import (
	"path/filepath"
	"testing"

	"github.com/Jinnrry/pmail/models"
	_ "modernc.org/sqlite"
	"xorm.io/xorm"
)

func TestAuthVersionMigrationPreservesExistingUsers(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite", filepath.Join(t.TempDir(), "legacy.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()
	_, err = engine.Exec(`CREATE TABLE user (
		id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, account TEXT NOT NULL UNIQUE,
		name TEXT NOT NULL, password TEXT NOT NULL, disabled INTEGER NOT NULL DEFAULT 0,
		is_admin INTEGER NOT NULL DEFAULT 0)`)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := engine.Exec(`INSERT INTO user (account,name,password) VALUES ('existing','Existing user','test-hash')`); err != nil {
		t.Fatal(err)
	}
	// Database startup uses this same Sync2 call to add the version column.
	if err := engine.Sync2(&models.User{}); err != nil {
		t.Fatal(err)
	}
	var user models.User
	found, err := engine.Where("account=?", "existing").Get(&user)
	if err != nil || !found {
		t.Fatalf("existing user missing: %v", err)
	}
	if user.AuthVersion != 0 || user.Password != "test-hash" {
		t.Fatal("migration changed existing credentials")
	}
}
