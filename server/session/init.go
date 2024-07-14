package session

import (
	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/sqlite3store"
	"github.com/alexedwards/scs/v2"
	"pmail/config"
	"pmail/db"

	"time"
)

var Instance *scs.SessionManager

func Init() {
	Instance = scs.New()
	Instance.Lifetime = 7 * 24 * time.Hour
	// 使用db存储session数据，目前为了架构简单，
	// 暂不引入redis存储，如果日后性能存在瓶颈，可以将session迁移到redis

	switch config.Instance.DbType {
	case config.DBTypeMySQL:
		Instance.Store = mysqlstore.New(db.Instance.DB().DB)
	case config.DBTypeSQLite:
		Instance.Store = sqlite3store.New(db.Instance.DB().DB)
	case config.DBTypePostgres:
		Instance.Store = postgresstore.New(db.Instance.DB().DB)
	default:
		panic("Unsupported database type: " + config.Instance.DbType)
	}

}
