package db

import (
	"fmt"
	"github.com/Jinnrry/pmail/config"
	"github.com/Jinnrry/pmail/models"
	"github.com/Jinnrry/pmail/utils/context"
	"github.com/Jinnrry/pmail/utils/errors"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	_ "modernc.org/sqlite"
	"xorm.io/xorm"
)

var Instance *xorm.Engine

func Init(version string) error {
	dsn := config.Instance.DbDSN
	var err error

	switch config.Instance.DbType {
	case "mysql":
		Instance, err = xorm.NewEngine("mysql", dsn)
		Instance.SetMaxOpenConns(100)
		Instance.SetMaxIdleConns(10)
	case "sqlite":
		Instance, err = xorm.NewEngine("sqlite", dsn)
		Instance.SetMaxOpenConns(1)
		Instance.SetMaxIdleConns(1)
	case "postgres":
		Instance, err = xorm.NewEngine("postgres", dsn)
		Instance.SetMaxOpenConns(100)
		Instance.SetMaxIdleConns(10)
	default:
		return errors.New("Database Type Error!")
	}
	if err != nil {
		return errors.Wrap(err)
	}

	Instance.ShowSQL(false)
	// 同步表结构
	syncTables()

	// 更新历史数据
	fixHistoryData()

	// 在数据库中记录程序版本
	var v models.Version
	_, err = Instance.Get(&v)
	if err != nil {
		panic(err)
	}

	if version != "" && v.Info != version {
		v.Info = version
		Instance.Update(&v)
	}

	if config.Instance.LogLevel == "debug" {
		Instance.ShowSQL(true)
	}

	return nil
}

func WithContext(ctx *context.Context, sql string) string {
	if ctx != nil {
		logId := ctx.GetValue(context.LogID)
		return fmt.Sprintf("/* %s */ %s", logId, sql)
	}
	return sql
}

func syncTables() {
	err := Instance.Sync2(&models.User{})
	if err != nil {
		panic(err)
	}
	err = Instance.Sync2(&models.Email{})
	if err != nil {
		panic(err)
	}
	err = Instance.Sync2(&models.Group{})
	if err != nil {
		panic(err)
	}
	err = Instance.Sync2(&models.Rule{})
	if err != nil {
		panic(err)
	}
	err = Instance.Sync2(&models.Sessions{})
	if err != nil {
		panic(err)
	}
	err = Instance.Sync2(&models.UserEmail{})
	if err != nil {
		panic(err)
	}
	err = Instance.Sync2(&models.Version{})
	if err != nil {
		panic(err)
	}
}

func fixHistoryData() {
	var ueNum int
	_, err := Instance.Table(&models.UserEmail{}).Select("count(1)").Get(&ueNum)
	if err != nil {
		panic(err)
	}
	if ueNum > 0 {
		return
	}

	// 只有一个管理员用户
	var user []models.User
	err = Instance.Table(&models.User{}).OrderBy("id asc").Find(&user)
	if err != nil {
		panic(err)
	}

	// 只有一个账号，且不是管理员账号，将账号提权为管理员
	if len(user) == 1 && user[0].IsAdmin == 0 {
		u := user[0]
		u.IsAdmin = 1
		_, err = Instance.Update(&u)
		if err != nil {
			panic(err)
		}
	}

	if len(user) != 1 {
		return
	}

	// 以前有邮件
	var emails []*models.Email
	err = Instance.Table(&models.Email{}).Select("id,status").OrderBy("id asc").Find(&emails)
	if err != nil {
		panic(err)
	}
	if len(emails) == 0 {
		return
	}

	log.Infof("Sync History Data！Please Wait！")

	// 把以前的邮件，全部分到管理员账号下面去
	for _, email := range emails {
		ue := models.UserEmail{
			UserID:  user[0].ID,
			EmailID: email.Id,
			Status:  email.Status,
		}
		_, err = Instance.Insert(&ue)
		if err != nil {
			log.Errorf("SQL Error: %v", err)
		}
	}
	log.Infof("Sync History Data Finished. Num: %d", len(emails))

}
