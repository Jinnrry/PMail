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
	"pmail/services/rule"
	"pmail/utils/address"
	"pmail/utils/array"
	"pmail/utils/context"
)

func GetRule(ctx *context.Context, w http.ResponseWriter, req *http.Request) {
	res := rule.GetAllRules(ctx)
	response.NewSuccessResponse(res).FPrint(w)
}

func UpsertRule(ctx *context.Context, w http.ResponseWriter, req *http.Request) {

	requestBody, err := io.ReadAll(req.Body)
	if err != nil {
		log.WithContext(ctx).Errorf("ReadError:%v", err)
		return
	}

	var data *dto.Rule
	err = json.Unmarshal(requestBody, &data)
	if err != nil {
		response.NewErrorResponse(response.ParamsError, "params error", err).FPrint(w)
		return
	}

	if data.Action == dto.FORWARD && !address.IsValidEmailAddress(data.Params) {

		response.NewErrorResponse(response.ParamsError, "ParamsError error", i18n.GetText(ctx.Lang, "invalid_email_address")).FPrint(w)
		return
	}

	for _, r := range data.Rules {
		if !array.InArray(r.Field, []string{"From", "Subject", "To", "Cc", "Text", "Html", "Content"}) {
			response.NewErrorResponse(response.ParamsError, "ParamsError error", "params error! Rule Field Error!").FPrint(w)
			return
		}
	}

	err = data.Encode().Save(ctx)
	if err != nil {
		response.NewErrorResponse(response.ServerError, "server error", err).FPrint(w)
		return
	}
	response.NewSuccessResponse("succ").FPrint(w)
}

type delRuleReq struct {
	Id int `json:"id"`
}

func DelRule(ctx *context.Context, w http.ResponseWriter, req *http.Request) {
	requestBody, err := io.ReadAll(req.Body)
	if err != nil {
		log.WithContext(ctx).Errorf("ReadError:%v", err)
		return
	}

	var data delRuleReq
	err = json.Unmarshal(requestBody, &data)
	if err != nil {
		response.NewErrorResponse(response.ParamsError, "params error", err).FPrint(w)
		return
	}

	if data.Id <= 0 {
		response.NewErrorResponse(response.ParamsError, "params error", "id is empty").FPrint(w)
		return
	}

	_, err = db.Instance.Exec(db.WithContext(ctx, "delete from rule where id =? and user_id =?"), data.Id, ctx.UserID)
	if err != nil {
		response.NewErrorResponse(response.ServerError, "unknown error", err).FPrint(w)
		return
	}

	response.NewSuccessResponse("succ").FPrint(w)
}
