package controllers

import (
	"encoding/json"
	"github.com/Jinnrry/pmail/dto"
	"github.com/Jinnrry/pmail/dto/response"
	"github.com/Jinnrry/pmail/i18n"
	"github.com/Jinnrry/pmail/models"
	"github.com/Jinnrry/pmail/services/group"
	"github.com/Jinnrry/pmail/utils/array"
	"github.com/Jinnrry/pmail/utils/context"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
)

func GetUserGroupList(ctx *context.Context, w http.ResponseWriter, req *http.Request) {
	defaultGroup := []*models.Group{
		{models.INBOX, i18n.GetText(ctx.Lang, "inbox"), 0, 0, "/"},     // 收件箱
		{models.Junk, i18n.GetText(ctx.Lang, "junk"), 0, 0, "/"},       //垃圾邮件
		{models.Deleted, i18n.GetText(ctx.Lang, "deleted"), 0, 0, "/"}, //已删除
	}

	infos := group.GetGroupList(ctx)

	response.NewSuccessResponse(append(defaultGroup, infos...)).FPrint(w)
}

func GetUserGroup(ctx *context.Context, w http.ResponseWriter, req *http.Request) {

	retData := []*group.GroupItem{
		{
			Label: i18n.GetText(ctx.Lang, "all_email"),
			Children: []*group.GroupItem{
				{
					Label: i18n.GetText(ctx.Lang, "inbox"),
					Tag:   dto.SearchTag{Type: 0, Status: -1, GroupId: 0}.ToString(),
				},
				{
					Label: i18n.GetText(ctx.Lang, "outbox"),
					Tag:   dto.SearchTag{Type: 1, Status: -1}.ToString(),
				},
				{
					Label: i18n.GetText(ctx.Lang, "sketch"),
					Tag:   dto.SearchTag{Type: 0, Status: 4}.ToString(),
				},
				{
					Label: i18n.GetText(ctx.Lang, "junk"),
					Tag:   dto.SearchTag{Type: -1, Status: 5}.ToString(),
				},
				{
					Label: i18n.GetText(ctx.Lang, "deleted"),
					Tag:   dto.SearchTag{Type: -1, Status: 3}.ToString(),
				},
			},
		},
	}

	retData = array.Merge(retData, group.GetGroupInfoList(ctx))

	response.NewSuccessResponse(retData).FPrint(w)
}

type addGroupRequest struct {
	Name     string `json:"name"`
	ParentId int    `json:"parent_id"`
}

func AddGroup(ctx *context.Context, w http.ResponseWriter, req *http.Request) {
	var reqData *addGroupRequest
	reqBytes, err := io.ReadAll(req.Body)
	if err != nil {
		log.WithContext(ctx).Errorf("%+v", err)
	}
	err = json.Unmarshal(reqBytes, &reqData)
	if err != nil {
		log.WithContext(ctx).Errorf("%+v", err)
	}

	newGroup, err := group.CreateGroup(ctx, reqData.Name, reqData.ParentId)

	if err != nil {
		response.NewErrorResponse(response.ServerError, "DBError", err.Error()).FPrint(w)
		return
	}

	response.NewSuccessResponse(newGroup.ID).FPrint(w)
}

type delGroupRequest struct {
	Id int `json:"id"`
}

func DelGroup(ctx *context.Context, w http.ResponseWriter, req *http.Request) {
	var reqData *delGroupRequest
	reqBytes, err := io.ReadAll(req.Body)
	if err != nil {
		log.WithContext(ctx).Errorf("%+v", err)
	}
	err = json.Unmarshal(reqBytes, &reqData)
	if err != nil {
		log.WithContext(ctx).Errorf("%+v", err)
	}
	succ, err := group.DelGroup(ctx, reqData.Id)

	if err != nil {
		response.NewErrorResponse(response.ServerError, "DBError", err.Error()).FPrint(w)
		return
	}
	response.NewSuccessResponse(succ).FPrint(w)
}
