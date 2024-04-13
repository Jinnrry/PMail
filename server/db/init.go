package db

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	_ "modernc.org/sqlite"
	"pmail/config"
	"pmail/utils/context"
	"pmail/utils/errors"
	"xorm.io/xorm"
)

var Instance *xorm.Engine

func Init() error {
	dsn := config.Instance.DbDSN
	var err error

	switch config.Instance.DbType {
	case "mysql":
		Instance, err = xorm.NewEngine("mysql", dsn)
	case "sqlite":
		Instance, err = xorm.NewEngine("sqlite", dsn)
	default:
		return errors.New("Database Type Error!")
	}
	if err != nil {
		return errors.Wrap(err)
	}
	Instance.SetMaxOpenConns(100)
	Instance.SetMaxIdleConns(10)

	return nil
}

func WithContext(ctx *context.Context, sql string) string {
	if ctx != nil {
		logId := ctx.GetValue(context.LogID)
		return fmt.Sprintf("/* %s */ %s", logId, sql)
	}
	return sql
}
