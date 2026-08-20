package db

import (
	stderrors "errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode"

	log "github.com/sirupsen/logrus"
	"modernc.org/sqlite"
	"xorm.io/xorm"
)

const (
	sqliteBusyTimeoutMilliseconds = 1000
	transactionAttempts           = 3
	transactionInitialBackoff     = 50 * time.Millisecond
)

// Transaction runs fn atomically. SQLite BUSY and LOCKED failures retry the
// entire transaction, so fn must contain only work that is safe to repeat.
func Transaction(engine *xorm.Engine, fn func(*xorm.Session) error) error {
	return transactionWithRetry(engine, transactionAttempts, transactionInitialBackoff, fn)
}

func transactionWithRetry(engine *xorm.Engine, attempts int, initialBackoff time.Duration, fn func(*xorm.Session) error) error {
	if attempts < 1 {
		attempts = 1
	}

	var err error
	for attempt := 1; attempt <= attempts; attempt++ {
		var retrySafe bool
		err, retrySafe = runTransactionAttempt(engine, fn)
		if err == nil {
			return nil
		}
		if !retrySafe || engine.DriverName() != "sqlite" || !IsSQLiteBusyOrLocked(err) || attempt == attempts {
			return err
		}
		time.Sleep(initialBackoff << (attempt - 1))
	}
	return err
}

func runTransaction(engine *xorm.Engine, fn func(*xorm.Session) error) error {
	err, _ := runTransactionAttempt(engine, fn)
	return err
}

func runTransactionAttempt(engine *xorm.Engine, fn func(*xorm.Session) error) (error, bool) {
	session := engine.NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		return fmt.Errorf("begin transaction: %w", err), true
	}
	if err := fn(session); err != nil {
		if rollbackErr := session.Rollback(); rollbackErr != nil {
			return stderrors.Join(err, fmt.Errorf("rollback transaction: %w", rollbackErr)), false
		}
		return err, true
	}
	if err := session.Commit(); err != nil {
		// modernc.org/sqlite v1.46.1 and newer roll back internally when COMMIT
		// fails, leaving the connection clean and the transaction safe to retry.
		return fmt.Errorf("commit transaction: %w", err), true
	}
	return nil, false
}

// IsSQLiteBusyOrLocked classifies SQLite errors by result code, including
// extended BUSY and LOCKED codes. Error message text is intentionally ignored.
func IsSQLiteBusyOrLocked(err error) bool {
	var sqliteErr *sqlite.Error
	if !stderrors.As(err, &sqliteErr) {
		return false
	}
	primaryCode := sqliteErr.Code() & 0xff
	return primaryCode == 5 || primaryCode == 6
}

func sqliteDSNWithReliabilityOptions(dsn string) (string, error) {
	dsnPrefix, rawQuery, hasQuery := strings.Cut(dsn, "?")
	query, err := url.ParseQuery(rawQuery)
	if err != nil {
		return "", fmt.Errorf("parse SQLite DSN: %w", err)
	}

	pragmas := query["_pragma"]
	timeoutIndex := -1
	for index, pragma := range pragmas {
		if sqlitePragmaName(pragma) == "busy_timeout" {
			if timeoutIndex >= 0 {
				return "", fmt.Errorf("parse SQLite DSN: multiple busy_timeout pragmas are ambiguous")
			}
			timeoutIndex = index
		}
	}
	if timeoutIndex >= 0 {
		requested, err := sqliteBusyTimeoutValue(pragmas[timeoutIndex])
		if err != nil {
			return "", fmt.Errorf("parse SQLite busy_timeout pragma %q: %w", pragmas[timeoutIndex], err)
		}
		if requested <= sqliteBusyTimeoutMilliseconds {
			return dsn, nil
		}

		log.Warnf(
			"SQLite busy_timeout requested %dms exceeds the %dms maximum; using %dms",
			requested,
			sqliteBusyTimeoutMilliseconds,
			sqliteBusyTimeoutMilliseconds,
		)
		pragmas[timeoutIndex] = fmt.Sprintf("busy_timeout(%d)", sqliteBusyTimeoutMilliseconds)
		query["_pragma"] = pragmas
		return dsnPrefix + "?" + query.Encode(), nil
	}

	separator := "?"
	if hasQuery {
		separator = "&"
		if strings.HasSuffix(dsn, "?") || strings.HasSuffix(dsn, "&") {
			separator = ""
		}
	}
	option := url.Values{"_pragma": {fmt.Sprintf("busy_timeout(%d)", sqliteBusyTimeoutMilliseconds)}}.Encode()
	return dsn + separator + option, nil
}

func sqliteBusyTimeoutValue(pragma string) (int64, error) {
	const name = "busy_timeout"

	normalized := strings.TrimSpace(pragma)
	if len(normalized) < len(name) || !strings.EqualFold(normalized[:len(name)], name) {
		return 0, fmt.Errorf("not a busy_timeout pragma")
	}
	remainder := strings.TrimSpace(normalized[len(name):])

	var valueText string
	switch {
	case strings.HasPrefix(remainder, "("):
		if !strings.HasSuffix(remainder, ")") {
			return 0, fmt.Errorf("expected busy_timeout(<milliseconds>)")
		}
		valueText = strings.TrimSpace(remainder[1 : len(remainder)-1])
	case strings.HasPrefix(remainder, "="):
		valueText = strings.TrimSpace(remainder[1:])
	default:
		return 0, fmt.Errorf("expected busy_timeout(<milliseconds>) or busy_timeout = <milliseconds>")
	}
	if valueText == "" {
		return 0, fmt.Errorf("timeout value is empty")
	}

	timeout, err := strconv.ParseInt(valueText, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid timeout value %q: %w", valueText, err)
	}
	return timeout, nil
}

func sqlitePragmaName(pragma string) string {
	normalized := strings.ToLower(strings.TrimSpace(pragma))
	if end := strings.IndexFunc(normalized, func(character rune) bool {
		return character == '(' || character == '=' || unicode.IsSpace(character)
	}); end >= 0 {
		normalized = normalized[:end]
	}
	return normalized
}
