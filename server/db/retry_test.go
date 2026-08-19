package db

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"modernc.org/sqlite"
	"xorm.io/xorm"
)

func TestSQLiteDSNOptions(t *testing.T) {
	tests := []struct {
		name            string
		dsn             string
		wantDSN         string
		wantBusyTimeout string
		wantJournalMode string
		wantTxLock      string
	}{
		{
			name:            "plain filename gets default timeout",
			dsn:             "/tmp/pmail.db",
			wantBusyTimeout: "busy_timeout(1000)",
		},
		{
			name:            "timeout at cap is preserved byte for byte",
			dsn:             "file:pmail.db?_txlock=deferred&_pragma=BUSY_TIMEOUT%20%3D%201000&mode=memory",
			wantDSN:         "file:pmail.db?_txlock=deferred&_pragma=BUSY_TIMEOUT%20%3D%201000&mode=memory",
			wantBusyTimeout: "BUSY_TIMEOUT = 1000",
			wantTxLock:      "deferred",
		},
		{
			name:            "explicit timeout is capped and other options are preserved",
			dsn:             "file:pmail.db?mode=memory&_pragma=journal_mode(WAL)&_pragma=busy_timeout%20%3D%20600000&_txlock=deferred",
			wantBusyTimeout: "busy_timeout(1000)",
			wantJournalMode: "journal_mode(WAL)",
			wantTxLock:      "deferred",
		},
		{
			name:            "negative timeout is preserved byte for byte",
			dsn:             "file:pmail.db?cache=shared&_pragma=%20busy_timeout%20%28%20-1%20%29%20",
			wantDSN:         "file:pmail.db?cache=shared&_pragma=%20busy_timeout%20%28%20-1%20%29%20",
			wantBusyTimeout: " busy_timeout ( -1 ) ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sqliteDSNWithReliabilityOptions(tt.dsn)
			if err != nil {
				t.Fatalf("sqliteDSNWithReliabilityOptions() error = %v", err)
			}
			if tt.wantDSN != "" && got != tt.wantDSN {
				t.Errorf("sqliteDSNWithReliabilityOptions() = %q, want byte-for-byte %q", got, tt.wantDSN)
			}

			queryText := ""
			if _, after, ok := strings.Cut(got, "?"); ok {
				queryText = after
			}
			query, err := url.ParseQuery(queryText)
			if err != nil {
				t.Fatalf("ParseQuery(%q): %v", queryText, err)
			}
			if got := query.Get("_txlock"); got != tt.wantTxLock {
				t.Errorf("_txlock = %q, want %q", got, tt.wantTxLock)
			}
			if !containsString(query["_pragma"], tt.wantBusyTimeout) {
				t.Errorf("_pragma = %q, want %q", query["_pragma"], tt.wantBusyTimeout)
			}
			if got := countPragma(query["_pragma"], "busy_timeout"); got != 1 {
				t.Errorf("busy_timeout pragma count = %d, want 1: %q", got, query["_pragma"])
			}
			if tt.wantJournalMode != "" && !containsString(query["_pragma"], tt.wantJournalMode) {
				t.Errorf("_pragma = %q, want preserved %q", query["_pragma"], tt.wantJournalMode)
			}
		})
	}
}

func TestSQLiteDSNRejectsInvalidBusyTimeout(t *testing.T) {
	tests := []struct {
		name string
		dsn  string
	}{
		{
			name: "duplicate timeout pragmas",
			dsn:  "file:pmail.db?_pragma=busy_timeout(10)&_pragma=BUSY_TIMEOUT%20%3D%2020",
		},
		{
			name: "missing timeout value",
			dsn:  "file:pmail.db?_pragma=busy_timeout",
		},
		{
			name: "nonnumeric timeout",
			dsn:  "file:pmail.db?_pragma=busy_timeout(forever)",
		},
		{
			name: "timeout with trailing junk",
			dsn:  "file:pmail.db?_pragma=busy_timeout(10)%20milliseconds",
		},
		{
			name: "overflowing timeout",
			dsn:  "file:pmail.db?_pragma=busy_timeout(9223372036854775808)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := sqliteDSNWithReliabilityOptions(tt.dsn); err == nil {
				t.Fatalf("sqliteDSNWithReliabilityOptions(%q) error = nil, want error", tt.dsn)
			}
		})
	}
}

func TestSQLiteDSNWarnsWhenBusyTimeoutIsCapped(t *testing.T) {
	logger := log.StandardLogger()
	originalOutput := logger.Out
	originalLevel := logger.GetLevel()
	var output bytes.Buffer
	logger.SetOutput(&output)
	logger.SetLevel(log.WarnLevel)
	t.Cleanup(func() {
		logger.SetOutput(originalOutput)
		logger.SetLevel(originalLevel)
	})

	if _, err := sqliteDSNWithReliabilityOptions("file:pmail.db?_pragma=busy_timeout(600000)"); err != nil {
		t.Fatalf("sqliteDSNWithReliabilityOptions() error = %v", err)
	}
	for _, want := range []string{"600000ms", "1000ms"} {
		if !strings.Contains(output.String(), want) {
			t.Errorf("warning = %q, want it to contain %q", output.String(), want)
		}
	}
}

func TestSQLiteBusyTimeoutAppliesToEveryConnection(t *testing.T) {
	dsn, err := sqliteDSNWithReliabilityOptions(filepath.Join(t.TempDir(), "connections.db"))
	if err != nil {
		t.Fatal(err)
	}
	database, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	database.SetMaxOpenConns(2)

	first, err := database.Conn(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	second, err := database.Conn(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()

	for index, connection := range []*sql.Conn{first, second} {
		var timeout int
		if err = connection.QueryRowContext(context.Background(), "PRAGMA busy_timeout").Scan(&timeout); err != nil {
			t.Fatalf("connection %d: %v", index+1, err)
		}
		if timeout != sqliteBusyTimeoutMilliseconds {
			t.Errorf("connection %d busy_timeout = %d, want %d", index+1, timeout, sqliteBusyTimeoutMilliseconds)
		}
	}
}

func TestIsSQLiteBusyOrLockedUsesDriverCode(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite", filepath.Join(t.TempDir(), "busy.db")+"?_pragma=busy_timeout(1)")
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()
	if _, err = engine.Exec("CREATE TABLE audit (id INTEGER PRIMARY KEY, value TEXT)"); err != nil {
		t.Fatal(err)
	}

	busyErr := sqliteBusyError(t, engine.DataSourceName())
	var sqliteErr *sqlite.Error
	if !errors.As(busyErr, &sqliteErr) {
		t.Fatalf("error type = %T, want *sqlite.Error", busyErr)
	}
	if !IsSQLiteBusyOrLocked(fmt.Errorf("wrapped: %w", busyErr)) {
		t.Fatalf("IsSQLiteBusyOrLocked(%v) = false", busyErr)
	}
	if IsSQLiteBusyOrLocked(errors.New("database is locked")) {
		t.Fatal("plain error text must not be classified as SQLITE_BUSY")
	}
}

func TestTransactionRetriesWholeSQLiteTransaction(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite", filepath.Join(t.TempDir(), "retry.db")+"?_pragma=busy_timeout(10)")
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()
	engine.SetMaxOpenConns(1)
	if _, err = engine.Exec("CREATE TABLE audit (id INTEGER PRIMARY KEY, value TEXT)"); err != nil {
		t.Fatal(err)
	}
	busyErr := sqliteBusyError(t, engine.DataSourceName())

	attempts := 0
	err = transactionWithRetry(engine, 3, time.Millisecond, func(session *xorm.Session) error {
		attempts++
		if attempts == 1 {
			return busyErr
		}
		_, insertErr := session.Exec("INSERT INTO audit(value) VALUES ('saved')")
		return insertErr
	})
	if err != nil {
		t.Fatalf("transactionWithRetry() error = %v", err)
	}
	if attempts != 2 {
		t.Fatalf("attempts = %d, want 2", attempts)
	}
	assertAuditRowCount(t, engine, 1)
}

func TestTransactionRetriesSQLiteBusyCommit(t *testing.T) {
	engine, blocker, readTx := newCommitBusyTest(t, 10)

	attempts := 0
	err := transactionWithRetry(engine, 2, 5*time.Millisecond, func(session *xorm.Session) error {
		attempts++
		if attempts == 2 {
			if rollbackErr := readTx.Rollback(); rollbackErr != nil {
				return fmt.Errorf("release read transaction: %w", rollbackErr)
			}
		}
		_, insertErr := session.Exec("INSERT INTO audit(value) VALUES ('saved')")
		return insertErr
	})
	_ = blocker.Close()
	if err != nil {
		t.Fatalf("transactionWithRetry() error = %v", err)
	}
	if attempts != 2 {
		t.Fatalf("attempts = %d, want 2", attempts)
	}
	assertAuditRowCount(t, engine, 2)
}

func TestPermanentSQLiteBusyCommitLeavesConnectionUsable(t *testing.T) {
	engine, blocker, readTx := newCommitBusyTest(t, 5)

	attempts := 0
	err := transactionWithRetry(engine, 2, time.Millisecond, func(session *xorm.Session) error {
		attempts++
		_, insertErr := session.Exec("INSERT INTO audit(value) VALUES ('rolled back')")
		return insertErr
	})
	if err == nil || !IsSQLiteBusyOrLocked(err) {
		t.Fatalf("transactionWithRetry() error = %v, want SQLITE_BUSY", err)
	}
	if attempts != 2 {
		t.Fatalf("attempts = %d, want 2", attempts)
	}
	if err = readTx.Rollback(); err != nil {
		t.Fatal(err)
	}
	_ = blocker.Close()

	err = runTransaction(engine, func(session *xorm.Session) error {
		_, insertErr := session.Exec("INSERT INTO audit(value) VALUES ('after busy')")
		return insertErr
	})
	if err != nil {
		t.Fatalf("transaction after BUSY = %v", err)
	}
	assertAuditRowCount(t, engine, 2)
}

func TestTransactionDoesNotClassifyBusyText(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite", filepath.Join(t.TempDir(), "text.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()

	attempts := 0
	wantErr := errors.New("database is locked")
	err = transactionWithRetry(engine, 3, time.Millisecond, func(*xorm.Session) error {
		attempts++
		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("transactionWithRetry() error = %v, want %v", err, wantErr)
	}
	if attempts != 1 {
		t.Fatalf("attempts = %d, want 1", attempts)
	}
}

func newCommitBusyTest(t *testing.T, timeoutMilliseconds int) (*xorm.Engine, *sql.DB, *sql.Tx) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "commit-busy.db")
	dsn := fmt.Sprintf("%s?_pragma=busy_timeout(%d)", path, timeoutMilliseconds)
	engine, err := xorm.NewEngine("sqlite", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = engine.Close() })
	engine.SetMaxOpenConns(1)
	engine.SetMaxIdleConns(1)
	if _, err = engine.Exec("CREATE TABLE audit (id INTEGER PRIMARY KEY, value TEXT)"); err != nil {
		t.Fatal(err)
	}
	if _, err = engine.Exec("INSERT INTO audit(value) VALUES ('seed')"); err != nil {
		t.Fatal(err)
	}

	blocker, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatal(err)
	}
	readTx, err := blocker.Begin()
	if err != nil {
		_ = blocker.Close()
		t.Fatal(err)
	}
	var count int
	if err = readTx.QueryRow("SELECT count(*) FROM audit").Scan(&count); err != nil {
		_ = readTx.Rollback()
		_ = blocker.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = readTx.Rollback()
		_ = blocker.Close()
	})
	return engine, blocker, readTx
}

func sqliteBusyError(t *testing.T, dsn string) error {
	t.Helper()
	first, err := xorm.NewEngine("sqlite", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	second, err := xorm.NewEngine("sqlite", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()

	tx := first.NewSession()
	defer tx.Close()
	if err = tx.Begin(); err != nil {
		t.Fatal(err)
	}
	if _, err = tx.Exec("INSERT INTO audit(value) VALUES ('rolled back')"); err != nil {
		t.Fatal(err)
	}
	_, busyErr := second.Exec("INSERT INTO audit(value) VALUES ('blocked')")
	_ = tx.Rollback()
	if busyErr == nil {
		t.Fatal("expected SQLITE_BUSY")
	}
	return busyErr
}

func assertAuditRowCount(t *testing.T, engine *xorm.Engine, want int) {
	t.Helper()
	var count int
	if _, err := engine.SQL("SELECT count(*) FROM audit").Get(&count); err != nil {
		t.Fatal(err)
	}
	if count != want {
		t.Fatalf("audit rows = %d, want %d", count, want)
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func countPragma(values []string, name string) int {
	count := 0
	for _, value := range values {
		if sqlitePragmaName(value) == name {
			count++
		}
	}
	return count
}
