package group

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"pmail/db"
	"pmail/dto"
	"pmail/models"
	"pmail/utils/array"
	"pmail/utils/context"
	"pmail/utils/errors"
)

type GroupItem struct {
	Id       int          `json:"id"`
	Label    string       `json:"label"`
	Tag      string       `json:"tag"`
	Children []*GroupItem `json:"children"`
}

func DelGroup(ctx *context.Context, groupId int) (bool, error) {
	allGroupIds := getAllChildId(ctx, groupId)
	allGroupIds = append(allGroupIds, groupId)

	// 开启一个事务
	trans, err := db.Instance.Begin()
	if err != nil {
		return false, errors.Wrap(err)
	}

	res, err := trans.Exec(db.WithContext(ctx, fmt.Sprintf("delete from `group` where id in (%s) and user_id =?", array.Join(allGroupIds, ","))), ctx.UserID)
	if err != nil {
		trans.Rollback()
		return false, errors.Wrap(err)
	}
	num, err := res.RowsAffected()
	if err != nil {
		trans.Rollback()
		return false, errors.Wrap(err)
	}

	_, err = trans.Exec(db.WithContext(ctx, fmt.Sprintf("update email set group_id=0 where group_id in (%s)", array.Join(allGroupIds, ","))))
	if err != nil {
		trans.Rollback()
		return false, errors.Wrap(err)
	}

	trans.Commit()

	return num > 0, nil
}

type id struct {
	Id int `db:"id"`
}

func getAllChildId(ctx *context.Context, rootId int) []int {
	var ids []id
	var ret []int
	db.Instance.Select(&ids, db.WithContext(ctx, "select id from `group` where parent_id=? and user_id=?"), rootId, ctx.UserID)
	for _, item := range ids {
		ret = array.Merge(ret, getAllChildId(ctx, item.Id))
		ret = append(ret, item.Id)
	}
	return ret
}

// GetGroupInfoList 获取全部的分组
func GetGroupInfoList(ctx *context.Context) []*GroupItem {
	return buildChildren(ctx, 0)
}

// MoveMailToGroup 将某封邮件移动到某个分组中
func MoveMailToGroup(ctx *context.Context, mailId []int, groupId int) bool {
	res, err := db.Instance.Exec(db.WithContext(ctx, fmt.Sprintf("update email set group_id=? where id in (%s)", array.Join(mailId, ","))), groupId)
	if err != nil {
		log.WithContext(ctx).Errorf("SQL Error:%+v", err)
		return false
	}
	rowNum, err := res.RowsAffected()
	if err != nil {
		log.WithContext(ctx).Errorf("SQL Error:%+v", err)
		return false
	}

	return rowNum > 0
}

func buildChildren(ctx *context.Context, parentId int) []*GroupItem {
	var ret []*GroupItem
	var rootGroup []*models.Group
	err := db.Instance.Select(&rootGroup, db.WithContext(ctx, "select * from `group` where parent_id=? and user_id=?"), parentId, ctx.UserID)

	if err != nil {
		log.WithContext(ctx).Errorf("SQL Error:%v", err)
	}

	for _, group := range rootGroup {
		ret = append(ret, &GroupItem{
			Id:       group.ID,
			Label:    group.Name,
			Tag:      dto.SearchTag{GroupId: group.ID, Status: -1, Type: -1}.ToString(),
			Children: buildChildren(ctx, group.ID),
		})
	}

	return ret

}

func GetGroupList(ctx *context.Context) []*models.Group {
	var ret []*models.Group
	db.Instance.Select(&ret, db.WithContext(ctx, "select * from `group` where user_id=?"), ctx.UserID)
	return ret
}
