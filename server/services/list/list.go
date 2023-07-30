package list

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"pmail/dto"
	"pmail/models"
	"pmail/mysql"
)

func GetEmailList(ctx *dto.Context, tag string, keyword string, offset, limit int) (emailList []*models.Email, total int) {

	querySQL, queryParams := genSQL(ctx, false, tag, keyword, offset, limit)
	counterSQL, counterParams := genSQL(ctx, true, tag, keyword, offset, limit)

	err := mysql.Instance.Select(&emailList, querySQL, queryParams...)
	if err != nil {
		log.Errorf("SQL ERROR: %s ,Error:%s", querySQL, err)
	}

	err = mysql.Instance.Get(&total, counterSQL, counterParams...)
	if err != nil {
		log.Errorf("SQL ERROR: %s ,Error:%s", querySQL, err)
	}

	return
}

func genSQL(ctx *dto.Context, counter bool, tag, keyword string, offset, limit int) (string, []any) {

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
	}

	if keyword != "" {
		sql += " and (subject like ? or text like ? )"
		sqlParams = append(sqlParams, "%"+keyword+"%", "%"+keyword+"%")
	}

	sql += " limit ? offset ?"
	sqlParams = append(sqlParams, limit, offset)

	return sql, sqlParams
}
