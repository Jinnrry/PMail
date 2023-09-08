package list

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"pmail/db"
	"pmail/dto"
	"pmail/models"
	"pmail/utils/context"
)

func GetEmailList(ctx *context.Context, tag string, keyword string, offset, limit int) (emailList []*models.Email, total int) {

	querySQL, queryParams := genSQL(ctx, false, tag, keyword, offset, limit)
	counterSQL, counterParams := genSQL(ctx, true, tag, keyword, offset, limit)

	err := db.Instance.Select(&emailList, db.WithContext(ctx, querySQL), queryParams...)
	if err != nil {
		log.Errorf("SQL ERROR: %s ,Error:%s", querySQL, err)
	}

	err = db.Instance.Get(&total, db.WithContext(ctx, counterSQL), counterParams...)
	if err != nil {
		log.Errorf("SQL ERROR: %s ,Error:%s", querySQL, err)
	}

	return
}

func genSQL(ctx *context.Context, counter bool, tag, keyword string, offset, limit int) (string, []any) {

	sql := "select * from email where 1=1 "
	if counter {
		sql = "select count(1) from email where 1=1 "
	}

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

	sql += " order by id desc limit ? offset ?"
	sqlParams = append(sqlParams, limit, offset)

	return sql, sqlParams
}
