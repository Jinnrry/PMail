package list

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"pmail/db"
	"pmail/dto"
	"pmail/models"
	"pmail/utils/context"
)

func GetEmailList(ctx *context.Context, tag string, keyword string, offset, limit int) (emailList []*models.Email, total int64) {

	querySQL, queryParams := genSQL(ctx, tag, keyword)

	total, err := db.Instance.Table("email").Where(querySQL, queryParams...).Desc("id").Limit(limit, offset).FindAndCount(&emailList)
	if err != nil {
		log.Errorf("SQL ERROR: %s ,Error:%s", querySQL, err)
	}

	return
}

func genSQL(ctx *context.Context, tag, keyword string) (string, []any) {

	sql := "1=1 "

	sqlParams := []any{}

	var tagInfo dto.SearchTag
	_ = json.Unmarshal([]byte(tag), &tagInfo)

	if tagInfo.Type != -1 {
		sql += " and type =? "
		sqlParams = append(sqlParams, tagInfo.Type)
	}

	if tagInfo.Status != -1 {
		sql += " and status =? "
		sqlParams = append(sqlParams, tagInfo.Status)
	} else {
		sql += " and status != 3"
	}

	if tagInfo.GroupId != -1 {
		sql += " and group_id=? "
		sqlParams = append(sqlParams, tagInfo.GroupId)
	}

	if keyword != "" {
		sql += " and (subject like ? or text like ? )"
		sqlParams = append(sqlParams, "%"+keyword+"%", "%"+keyword+"%")
	}

	return sql, sqlParams
}
