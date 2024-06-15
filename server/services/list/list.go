package list

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"pmail/db"
	"pmail/dto"
	"pmail/dto/response"
	"pmail/utils/context"
)

func GetEmailList(ctx *context.Context, tagInfo dto.SearchTag, keyword string, pop3List bool, offset, limit int) (emailList []*response.EmailResponseData, total int64) {
	return getList(ctx, tagInfo, keyword, pop3List, offset, limit)
}

func getList(ctx *context.Context, tagInfo dto.SearchTag, keyword string, pop3List bool, offset, limit int) (emailList []*response.EmailResponseData, total int64) {
	querySQL, queryParams := genSQL(ctx, false, tagInfo, keyword, pop3List, offset, limit)

	err := db.Instance.SQL(querySQL, queryParams...).Find(&emailList)
	if err != nil {
		log.WithContext(ctx).Errorf("SQL ERROR: %s ,Error:%s", querySQL, err)
	}

	totalSQL, totalParams := genSQL(ctx, true, tagInfo, keyword, pop3List, offset, limit)

	_, err = db.Instance.SQL(totalSQL, totalParams...).Get(&total)
	if err != nil {
		log.WithContext(ctx).Errorf("SQL ERROR: %s ,Error:%s", querySQL, err)
	}

	return emailList, total
}

func genSQL(ctx *context.Context, count bool, tagInfo dto.SearchTag, keyword string, pop3List bool, offset, limit int) (string, []any) {
	sqlParams := []any{ctx.UserID}
	sql := "select "

	if count {
		sql += `count(1) from email e left join user_email ue on e.id=ue.email_id where ue.user_id = ? `
	} else if pop3List {
		sql += `e.id,e.size from email e left join user_email ue on e.id=ue.email_id where ue.user_id = ? `
	} else {
		sql += `e.*,ue.is_read from email e left join user_email ue on e.id=ue.email_id where ue.user_id = ? `
	}

	if tagInfo.Status != -1 {
		sql += " and ue.status =? "
		sqlParams = append(sqlParams, tagInfo.Status)
	} else {
		sql += " and ue.status != 3"
	}

	if tagInfo.Type != -1 {
		sql += " and type =? "
		sqlParams = append(sqlParams, tagInfo.Type)
	}

	if tagInfo.GroupId != -1 {
		sql += " and ue.group_id=? "
		sqlParams = append(sqlParams, tagInfo.GroupId)
	}

	if keyword != "" {
		sql += " and (subject like ? or text like ? )"
		sqlParams = append(sqlParams, "%"+keyword+"%", "%"+keyword+"%")
	}

	if limit == 0 {
		limit = 10
	}

	sql += " order by e.id desc"

	if limit < 10000 {
		sql += fmt.Sprintf(" limit %d,%d ", offset, limit)
	}

	return sql, sqlParams

}

type statRes struct {
	Total int64
	Size  int64
}

// Stat 查询邮件总数和大小
func Stat(ctx *context.Context) (int64, int64) {
	sql := `select count(1) as total,sum(size) as size from email e left join user_email ue on e.id=ue.email_id where ue.user_id = ? and e.type = 0 and ue.status != 3`
	var ret statRes
	_, err := db.Instance.SQL(sql, ctx.UserID).Get(&ret)
	if err != nil {
		log.WithContext(ctx).Errorf("SQL ERROR: %s ,Error:%s", sql, err)
	}
	return ret.Total, ret.Size
}
