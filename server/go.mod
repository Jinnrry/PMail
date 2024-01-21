module pmail

go 1.21

require (
	github.com/Jinnrry/gopop v0.0.0-20231113115125-fbdf52ae39ea
	github.com/alexedwards/scs/mysqlstore v0.0.0-20230327161757-10d4299e3b24
	github.com/alexedwards/scs/sqlite3store v0.0.0-20230327161757-10d4299e3b24
	github.com/alexedwards/scs/v2 v2.5.1
	github.com/emersion/go-message v0.18.0
	github.com/emersion/go-msgauth v0.6.6
	github.com/emersion/go-smtp v0.20.1
	github.com/go-acme/lego/v4 v4.13.3
	github.com/go-sql-driver/mysql v1.7.1
	github.com/jmoiron/sqlx v1.3.5
	github.com/mileusna/spf v0.9.5
	github.com/sirupsen/logrus v1.9.3
	github.com/spf13/cast v1.5.1
	golang.org/x/crypto v0.10.0
	golang.org/x/text v0.14.0
	modernc.org/sqlite v1.24.0
)

replace github.com/alexedwards/scs/sqlite3store v0.0.0-20230327161757-10d4299e3b24 => github.com/Jinnrry/scs/sqlite3store v0.0.0-20230803080525-914f01e0d379

require (
	github.com/cenkalti/backoff/v4 v4.2.1 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/emersion/go-sasl v0.0.0-20200509203442-7bfe0ed36a21 // indirect
	github.com/emersion/go-textwrapper v0.0.0-20200911093747-65d896831594 // indirect
	github.com/go-jose/go-jose/v3 v3.0.0 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/mattn/go-sqlite3 v1.14.17 // indirect
	github.com/miekg/dns v1.1.55 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	golang.org/x/mod v0.11.0 // indirect
	golang.org/x/net v0.11.0 // indirect
	golang.org/x/sys v0.9.0 // indirect
	golang.org/x/tools v0.10.0 // indirect
	lukechampine.com/uint128 v1.2.0 // indirect
	modernc.org/cc/v3 v3.40.0 // indirect
	modernc.org/ccgo/v3 v3.16.13 // indirect
	modernc.org/libc v1.22.5 // indirect
	modernc.org/mathutil v1.5.0 // indirect
	modernc.org/memory v1.5.0 // indirect
	modernc.org/opt v0.1.3 // indirect
	modernc.org/strutil v1.1.3 // indirect
	modernc.org/token v1.0.1 // indirect
)
