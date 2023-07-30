package controllers

import (
	"net/http"
	"pmail/dto"
	"pmail/dto/response"
	"pmail/i18n"
)

type groupItem struct {
	Label    string       `json:"label"`
	Tag      string       `json:"tag"`
	Children []*groupItem `json:"children"`
}

func GetUserGroup(ctx *dto.Context, w http.ResponseWriter, req *http.Request) {

	retData := []*groupItem{
		{
			Label: i18n.GetText(ctx.Lang, "all_email"),
			Children: []*groupItem{
				{
					Label: i18n.GetText(ctx.Lang, "inbox"),
					Tag:   dto.SearchTag{Type: 0, Status: -1}.ToString(),
				},
				{
					Label: i18n.GetText(ctx.Lang, "outbox"),
					Tag:   dto.SearchTag{Type: 1, Status: 1}.ToString(),
				},
				{
					Label: i18n.GetText(ctx.Lang, "sketch"),
					Tag:   dto.SearchTag{Type: 1, Status: 0}.ToString(),
				},
			},
		},
	}

	response.NewSuccessResponse(retData).FPrint(w)
}
