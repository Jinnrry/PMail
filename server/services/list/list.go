package list

import (
	"fmt"
	"github.com/Jinnrry/pmail/db"
	"github.com/Jinnrry/pmail/dto"
	"github.com/Jinnrry/pmail/dto/response"
	"github.com/Jinnrry/pmail/models"
	"github.com/Jinnrry/pmail/utils/context"
	log "github.com/sirupsen/logrus"
	"strings"
)
import . "xorm.io/builder"

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
		sql += fmt.Sprintf(" LIMIT %d OFFSET %d ", limit, offset)
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

func GetEmailListByGroup(ctx *context.Context, groupName string, offset, limit int) []*response.EmailResponseData {
	if limit == 0 {
		limit = 1
	}

	var ret []*response.EmailResponseData
	var ue []*models.UserEmail
	switch groupName {
	case "INBOX":
		db.Instance.Table("user_email").Select("email_id,is_read").Where("user_id=? and status=0", ctx.UserID).Limit(limit, offset).Find(&ue)
	case "Sent Messages":
		db.Instance.Table("user_email").Select("email_id,is_read").Where("user_id=? and status=1", ctx.UserID).Limit(limit, offset).Find(&ue)
	case "Drafts":
		db.Instance.Table("user_email").Select("email_id,is_read").Where("user_id=? and status=4", ctx.UserID).Limit(limit, offset).Find(&ue)
	case "Deleted Messages":
		db.Instance.Table("user_email").Select("email_id,is_read").Where("user_id=? and status=3", ctx.UserID).Limit(limit, offset).Find(&ue)
	case "Junk":
		db.Instance.Table("user_email").Select("email_id,is_read").Where("user_id=? and status=5", ctx.UserID).Limit(limit, offset).Find(&ue)
	default:
		groupNames := strings.Split(groupName, "/")
		groupName = groupNames[len(groupNames)-1]

		var group models.Group
		db.Instance.Table("group").Where("user_id=? and name=?", ctx.UserID, groupName).Get(&group)
		if group.ID == 0 {
			return ret
		}
		db.Instance.Table("user_email").Select("email_id,is_read").Where("user_id=? and group_id = ?", ctx.UserID, group.ID).Limit(limit, offset).Find(&ue)
	}

	ueMap := map[int]*models.UserEmail{}
	var emailIds []int
	for _, email := range ue {
		ueMap[email.EmailID] = email
		emailIds = append(emailIds, email.EmailID)
	}

	_ = db.Instance.Table("email").Select("*").Where(Eq{"id": emailIds}).Find(&ret)
	for i, data := range ret {
		ret[i].IsRead = ueMap[data.Id].IsRead
	}

	return ret
}
