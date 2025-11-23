package group

import (
	errors2 "errors"
	"fmt"
	"github.com/Jinnrry/pmail/consts"
	"github.com/Jinnrry/pmail/db"
	"github.com/Jinnrry/pmail/dto"
	"github.com/Jinnrry/pmail/models"
	"github.com/Jinnrry/pmail/services/del_email"
	"github.com/Jinnrry/pmail/utils/array"
	"github.com/Jinnrry/pmail/utils/context"
	"github.com/Jinnrry/pmail/utils/errors"
	log "github.com/sirupsen/logrus"
	"strings"
	"xorm.io/builder"
)

type GroupItem struct {
	Id       int          `json:"id"`
	Label    string       `json:"label"`
	Tag      string       `json:"tag"`
	Children []*GroupItem `json:"children"`
}

func CreateGroup(ctx *context.Context, name string, parentId int) (*models.Group, error) {
	// 先查询是否存在
	var group models.Group
	db.Instance.Table("group").Where("name = ? and user_id = ?", name, ctx.UserID).Get(&group)
	if group.ID > 0 {
		return &group, nil
	}
	group.Name = name
	group.ParentId = parentId
	group.UserId = ctx.UserID
	group.FullPath = getLayerName(ctx, &group, true)

	_, err := db.Instance.Insert(&group)
	return &group, err
}

func Rename(ctx *context.Context, oldName, newName string) error {
	oldGroupInfo, err := GetGroupByName(ctx, oldName)
	if err != nil {
		return err
	}
	if oldGroupInfo == nil || oldGroupInfo.ID == 0 {
		return errors2.New("group not found")
	}
	oldGroupInfo.Name = newName
	oldGroupInfo.FullPath = getLayerName(ctx, oldGroupInfo, true)
	_, err = db.Instance.ID(oldGroupInfo.ID).Update(oldGroupInfo)
	return err
}

func GetGroupByName(ctx *context.Context, name string) (*models.Group, error) {
	var group models.Group
	db.Instance.Table("group").Where("name = ? and user_id = ?", name, ctx.UserID).Get(&group)

	return &group, nil
}

func GetGroupByFullPath(ctx *context.Context, fullPath string) (*models.Group, error) {
	var group models.Group
	_, err := db.Instance.Table("group").Where("full_path = ? and user_id = ?", fullPath, ctx.UserID).Get(&group)

	return &group, err
}

func DelGroup(ctx *context.Context, groupId int) (bool, error) {
	allGroupIds := getAllChildId(ctx, groupId)
	allGroupIds = append(allGroupIds, groupId)

	// 开启一个事务
	trans := db.Instance.NewSession()

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

	_, err = trans.Exec(db.WithContext(ctx, fmt.Sprintf("update user_email set group_id=0 where group_id in (%s)", array.Join(allGroupIds, ","))))
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
	err := db.Instance.Table("group").Where("parent_id=? and user_id=?", rootId, ctx.UserID).Find(&ids)
	if err != nil {
		log.WithContext(ctx).Errorf("getAllChildId err: %v", err)
	}
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
	res, err := db.Instance.Exec(db.WithContext(ctx,
		fmt.Sprintf("update user_email set group_id=? where email_id in (%s) and user_id =?", array.Join(mailId, ","))),
		groupId, ctx.UserID)
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
	err := db.Instance.Table("group").Where("parent_id=? and user_id=?", parentId, ctx.UserID).Find(&rootGroup)

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
	db.Instance.Table("group").Where("user_id=?", ctx.UserID).Find(&ret)
	return ret
}

func hasChildren(ctx *context.Context, id int) bool {
	var parent []*models.Group
	db.Instance.Table("group").Where("parent_id=?", id).Find(&parent)
	return len(parent) > 0
}

func getLayerName(ctx *context.Context, item *models.Group, allPath bool) string {
	if item.ParentId == 0 {
		return item.Name
	}
	var parent models.Group
	_, _ = db.Instance.Table("group").Where("id=?", item.ParentId).Get(&parent)
	if allPath {
		return getLayerName(ctx, &parent, allPath) + "/" + item.Name
	}
	return getLayerName(ctx, &parent, allPath)
}

func IsDefaultBox(box string) bool {
	return array.InArray(box, []string{"INBOX", "Sent Messages", "Drafts", "Deleted Messages", "Junk"})
}

func GetGroupStatus(ctx *context.Context, groupName string, params []string) (string, map[string]int) {
	retMap := map[string]int{}

	if !IsDefaultBox(groupName) {
		groupNames := strings.Split(groupName, "/")
		groupName = groupNames[len(groupNames)-1]

		var group models.Group
		db.Instance.Table("group").Where("user_id=? and name=?", ctx.UserID, groupName).Get(&group)
		if group.ID == 0 {
			ret := ""
			for _, param := range params {
				if ret != "" {
					ret += " "
				}
				retMap[param] = 0
				ret += fmt.Sprintf("%s %d", param, 0)
			}
			return fmt.Sprintf("(%s)", ret), retMap
		}
		ret := ""
		for _, param := range params {
			if ret != "" {
				ret += " "
			}
			var value int

			switch param {
			case "MESSAGES":
				db.Instance.Table("user_email").Select("count(1)").Where("group_id=?", group.ID).Get(&value)
			case "UIDNEXT":
				db.Instance.Table("user_email").Select("id").Where("group_id=?", group.ID).OrderBy("id desc").Get(&value)
				value += 1
			case "UIDVALIDITY":
				value = group.ID
			case "UNSEEN":
				db.Instance.Table("user_email").Select("count(1)").Where("group_id=? and is_read=0", group.ID).Get(&value)
			}
			retMap[param] = value
			ret += fmt.Sprintf("%s %d", param, value)
		}
		return fmt.Sprintf("(%s)", ret), retMap
	}

	ret := ""
	for _, param := range params {
		if ret != "" {
			ret += " "
		}
		var value int

		switch param {
		case "MESSAGES":
			value = getGroupNum(ctx, groupName, false)
		case "UIDNEXT":
			value = getNextUID(ctx, groupName)
		case "UIDVALIDITY":
			value = models.GroupNameToCode[groupName]
		case "UNSEEN":
			value = getGroupNum(ctx, groupName, true)
		default:
			continue
		}
		retMap[param] = value
		ret += fmt.Sprintf("%s %d", param, value)
	}
	if ret == "" {
		return "", retMap
	}

	return fmt.Sprintf("(%s)", ret), retMap

}

func getNextUID(ctx *context.Context, groupName string) int {
	var lastId int
	switch groupName {
	case "INBOX":
		db.Instance.Table("user_email").Select("id").Where("user_id=? and group_id=0 and status = 0", ctx.UserID).OrderBy("id desc").Get(&lastId)
	case "Sent Messages":
		db.Instance.Table("user_email").Select("id").Where("user_id=? and group_id=0 and status = 1", ctx.UserID).OrderBy("id desc").Get(&lastId)
	case "Drafts":
		db.Instance.Table("user_email").Select("id").Where("user_id=? and group_id=0 and status = 4", ctx.UserID).OrderBy("id desc").Get(&lastId)
	case "Deleted Messages":
		db.Instance.Table("user_email").Select("id").Where("user_id=? and status = 3", ctx.UserID).OrderBy("id desc").Get(&lastId)
	case "Junk":
		db.Instance.Table("user_email").Select("id").Where("user_id=? and group_id=0 and status = 5", ctx.UserID).OrderBy("id desc").Get(&lastId)
	}
	return lastId + 1
}

func getGroupNum(ctx *context.Context, groupName string, mustUnread bool) int {
	var count int
	switch groupName {
	case "INBOX":
		if mustUnread {
			db.Instance.Table("user_email").Select("count(1)").Where("user_id=? and status=0 and is_read=0", ctx.UserID).Get(&count)
		} else {
			db.Instance.Table("user_email").Select("count(1)").Where("user_id=? and status=0", ctx.UserID).Get(&count)
		}
	case "Sent Messages":
		if mustUnread {
			count = 0
		} else {
			db.Instance.Table("user_email").Select("count(1)").Where("user_id=? and status=1", ctx.UserID).Get(&count)
		}
	case "Drafts":
		if mustUnread {
			db.Instance.Table("user_email").Select("count(1)").Where("user_id=? and status=4 and is_read=0", ctx.UserID).Get(&count)
		} else {
			db.Instance.Table("user_email").Select("count(1)").Where("user_id=? and status=4", ctx.UserID).Get(&count)
		}
	case "Deleted Messages":
		if mustUnread {
			db.Instance.Table("user_email").Select("count(1)").Where("user_id=? and status=3 and is_read=0", ctx.UserID).Get(&count)
		} else {
			db.Instance.Table("user_email").Select("count(1)").Where("user_id=? and status=3", ctx.UserID).Get(&count)
		}
	case "Junk":
		if mustUnread {
			db.Instance.Table("user_email").Select("count(1)").Where("user_id=? and status=5 and is_read=0", ctx.UserID).Get(&count)
		} else {
			db.Instance.Table("user_email").Select("count(1)").Where("user_id=? and status=5", ctx.UserID).Get(&count)
		}
	}
	return count
}

func Move2DefaultBox(ctx *context.Context, mailIds []int, groupName string) error {
	switch groupName {
	case "Deleted Messages":
		err := del_email.DelEmail(ctx, mailIds, false)
		if err != nil {
			return err
		}
	case "INBOX":
		_, err := db.Instance.Table(&models.UserEmail{}).Where(builder.Eq{
			"user_id":  ctx.UserID,
			"email_id": mailIds,
		}).Update(map[string]interface{}{
			"status":   consts.EmailTypeReceive,
			"group_id": 0,
		})
		return err
	case "Sent Messages":
		_, err := db.Instance.Table(&models.UserEmail{}).Where(builder.Eq{
			"user_id":  ctx.UserID,
			"email_id": mailIds,
		}).Update(map[string]interface{}{
			"status":   consts.EmailStatusSent,
			"group_id": 0,
		})
		return err
	case "Drafts":
		_, err := db.Instance.Table(&models.UserEmail{}).Where(builder.Eq{
			"user_id":  ctx.UserID,
			"email_id": mailIds,
		}).Update(map[string]interface{}{
			"status":   consts.EmailStatusDrafts,
			"group_id": 0,
		})
		return err
	case "Junk":
		_, err := db.Instance.Table(&models.UserEmail{}).Where(builder.Eq{
			"user_id":  ctx.UserID,
			"email_id": mailIds,
		}).Update(map[string]interface{}{
			"status":   consts.EmailStatusJunk,
			"group_id": 0,
		})
		return err
	}
	return nil
}
