package db

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
	_ "modernc.org/sqlite"
	"pmail/config"
	"pmail/dto"
)

var Instance *sqlx.DB

func Init() {
	dsn := config.Instance.DbDSN
	var err error

	switch config.Instance.DbType {
	case "mysql":
		Instance, err = sqlx.Open("mysql", dsn)
	case "sqlite":
		Instance, err = sqlx.Open("sqlite", dsn)
	default:
		return
	}
	if err != nil {
		panic(err)
	}
	Instance.SetMaxOpenConns(100)
	Instance.SetMaxIdleConns(10)
	//showMySQLCharacterSet()
	checkTable()
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

func showMySQLCharacterSet() {
	var res []struct {
		Variable_name string `db:"Variable_name"`
		Value         string `db:"Value"`
	}
	err := Instance.Select(&res, "show variables like '%character%';")
	log.Debugf("%+v  %+v", res, err)

}

func testSlowLog() {
	var res []struct {
		Value string `db:"Value"`
	}
	err := Instance.Select(&res, "/* asddddasad */select /* this is test */ sleep(4) as Value")
	log.Debugf("%+v  %+v", res, err)

}
