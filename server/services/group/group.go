package group

import (
	"fmt"
	"github.com/Jinnrry/pmail/db"
	"github.com/Jinnrry/pmail/dto"
	"github.com/Jinnrry/pmail/models"
	"github.com/Jinnrry/pmail/utils/array"
	"github.com/Jinnrry/pmail/utils/context"
	"github.com/Jinnrry/pmail/utils/errors"
	"github.com/Jinnrry/pmail/utils/utf7"
	log "github.com/sirupsen/logrus"
	"strings"
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

func getLayerName(ctx *context.Context, item *models.Group) string {
	if item.ParentId == 0 {
		return utf7.Encode(item.Name)
	}
	var parent models.Group
	_, _ = db.Instance.Table("group").Where("id=?", item.ParentId).Get(&parent)
	return getLayerName(ctx, &parent) + "/" + utf7.Encode(item.Name)
}

func MatchGroup(ctx *context.Context, basePath, template string) []string {
	var groups []*models.Group
	var ret []string
	if basePath == "" {
		db.Instance.Table("group").Where("user_id=?", ctx.UserID).Find(&groups)
		ret = append(ret, `* LIST (\NoSelect \HasChildren) "/" "[PMail]"`)
		ret = append(ret, `* LIST (\HasNoChildren) "/" "INBOX"`)
		ret = append(ret, `* LIST (\HasNoChildren) "/" "Sent Messages"`)
		ret = append(ret, `* LIST (\HasNoChildren) "/" "Drafts"`)
		ret = append(ret, `* LIST (\HasNoChildren) "/" "Deleted Messages"`)
		ret = append(ret, `* LIST (\HasNoChildren) "/" "Junk"`)
	} else {
		var parent *models.Group
		db.Instance.Table("group").Where("user_id=? and name=?", ctx.UserID, basePath).Find(&groups)
		if parent != nil && parent.ID > 0 {
			db.Instance.Table("group").Where("user_id=? and parent_id=?", ctx.UserID, parent.ID).Find(&groups)
		}
	}
	for _, group := range groups {
		if hasChildren(ctx, group.ID) {
			ret = append(ret, fmt.Sprintf(`* LIST (\HasChildren) "/" "[PMail]/%s"`, getLayerName(ctx, group)))
		} else {
			ret = append(ret, fmt.Sprintf(`* LIST (\HasNoChildren) "/" "[PMail]/%s"`, getLayerName(ctx, group)))
		}
	}
	return ret
}

func GetGroupStatus(ctx *context.Context, groupName string, params []string) (string, map[string]int) {
	retMap := map[string]int{}

	if !array.InArray(groupName, []string{"INBOX", "Sent Messages", "Drafts", "Deleted Messages", "Junk"}) {
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
				db.Instance.Table("email").Select("count(1)").Get(&value)
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
			db.Instance.Table("email").Select("count(1)").Get(&value)
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
