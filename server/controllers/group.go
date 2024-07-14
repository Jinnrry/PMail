package controllers

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"pmail/db"
	"pmail/dto"
	"pmail/dto/response"
	"pmail/i18n"
	"pmail/models"
	"pmail/services/group"
	"pmail/utils/array"
	"pmail/utils/context"
)

func GetUserGroupList(ctx *context.Context, w http.ResponseWriter, req *http.Request) {
	infos := group.GetGroupList(ctx)
	response.NewSuccessResponse(infos).FPrint(w)
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
					Tag:   dto.SearchTag{Type: 1, Status: 0}.ToString(),
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

	var newGroup models.Group = models.Group{
		Name:     reqData.Name,
		ParentId: reqData.ParentId,
		UserId:   ctx.UserID,
	}

	_, err = db.Instance.Insert(&newGroup)
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
