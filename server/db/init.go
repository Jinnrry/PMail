package db

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
	_ "modernc.org/sqlite"
	"pmail/config"
	"pmail/dto"
	"pmail/utils/errors"
	"strings"
)

var Instance *sqlx.DB

func Init() error {
	dsn := config.Instance.DbDSN
	var err error

	switch config.Instance.DbType {
	case "mysql":
		Instance, err = sqlx.Open("mysql", dsn)
	case "sqlite":
		Instance, err = sqlx.Open("sqlite", dsn)
	default:
		return errors.New("Database Type Error!")
	}
	if err != nil {
		return errors.Wrap(err)
	}
	Instance.SetMaxOpenConns(100)
	Instance.SetMaxIdleConns(10)
	//showMySQLCharacterSet()
	checkTable()
	// 处理版本升级带来的数据表变更
	databaseUpdate()
	return nil
}

func WithContext(ctx *dto.Context, sql string) string {
	if ctx != nil {
		logId := ctx.GetValue(dto.LogID)
		return fmt.Sprintf("/* %s */ %s", logId, sql)
	}
	return sql
}

type tables struct {
	TablesInPmail string `db:"Tables_in_pmail"`
}

func checkTable() {
	var res []*tables

	var err error
	if config.Instance.DbType == "sqlite" {
		err = Instance.Select(&res, "select name as `Tables_in_pmail` from sqlite_master where type='table'")
	} else {
		err = Instance.Select(&res, "show tables")
	}
	if err != nil {
		panic(err)
	}
	existTable := map[string]struct{}{}
	for _, tableName := range res {
		existTable[tableName.TablesInPmail] = struct{}{}
	}

	for tableName, createSQL := range config.Instance.Tables {
		if _, ok := existTable[tableName]; !ok {
			_, err = Instance.Exec(createSQL)
			log.Infof("Create Table: %s", createSQL)
			if err != nil {
				panic(err)
			}

			if initData, ok := config.Instance.TablesInitData[tableName]; ok {
				_, err = Instance.Exec(initData)
				log.Infof("Init Table: %s", initData)
				if err != nil {
					panic(err)
				}
			}

		}
	}
}

type tableSQL struct {
	Table       string `db:"Table"`
	CreateTable string `db:"Create Table"`
}

func databaseUpdate() {
	// 检查email表是否有group id
	var err error
	var res []tableSQL
	if config.Instance.DbType == "sqlite" {
		err = Instance.Select(&res, "select sql as `Create Table` from sqlite_master where type='table' and tbl_name = 'email'")
	} else {
		err = Instance.Select(&res, "show create table `email`")
	}

	if err != nil {
		panic(err)
	}

	if len(res) > 0 && !strings.Contains(res[0].CreateTable, "group_id") {
		Instance.Exec("alter table email add group_id integer default 0 not null;")
	}

}
