package controllers

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"io"
	"math"
	"net/http"
	"pmail/config"
	"pmail/db"
	"pmail/dto/response"
	"pmail/models"
	"pmail/utils/array"
	"pmail/utils/context"
	"pmail/utils/password"
)

type userCreateRequest struct {
	Id       int    `json:"id"`
	Account  string `json:"account"`
	Username string `json:"username"`
	Password string `json:"password"`
	Disabled int    `json:"disabled"`
}

func CreateUser(ctx *context.Context, w http.ResponseWriter, req *http.Request) {

	if !ctx.IsAdmin {
		response.NewErrorResponse(response.NoAccessPrivileges, "No Access Privileges", "").FPrint(w)
		return
	}

	reqBytes, err := io.ReadAll(req.Body)
	if err != nil {
		log.Errorf("%+v", err)
	}
	var reqData userCreateRequest
	err = json.Unmarshal(reqBytes, &reqData)
	if err != nil {
		log.Errorf("%+v", err)
	}

	if reqData.Username == "" || reqData.Password == "" || reqData.Account == "" {
		response.NewErrorResponse(response.ParamsError, "Params Error", "").FPrint(w)
		return
	}

	var user models.User
	user.Name = reqData.Username
	user.Password = password.Encode(reqData.Password)
	user.Account = reqData.Account

	_, err = db.Instance.Insert(&user)
	if err != nil {
		response.NewErrorResponse(response.ServerError, err.Error(), "").FPrint(w)
		return
	}

	response.NewSuccessResponse(user).FPrint(w)
}

type userListRequest struct {
	CurrentPage int `json:"current_page"`
	PageSize    int `json:"page_size"`
}

func UserList(ctx *context.Context, w http.ResponseWriter, req *http.Request) {
	if !ctx.IsAdmin {
		response.NewErrorResponse(response.NoAccessPrivileges, "No Access Privileges", "").FPrint(w)
		return
	}

	reqBytes, err := io.ReadAll(req.Body)
	if err != nil {
		log.Errorf("%+v", err)
	}
	var reqData userListRequest
	err = json.Unmarshal(reqBytes, &reqData)
	if err != nil {
		log.Errorf("%+v", err)
	}

	offset := 0
	if reqData.CurrentPage >= 1 {
		offset = (reqData.CurrentPage - 1) * reqData.PageSize
	}

	if reqData.PageSize == 0 {
		reqData.PageSize = 15
	}

	var users []models.User

	totalNum, err := db.Instance.Table(&models.User{}).Limit(reqData.PageSize, offset).FindAndCount(&users)
	if err != nil {
		log.Errorf("%+v", err)
	}

	response.NewSuccessResponse(map[string]any{
		"current_page": reqData.CurrentPage,
		"total_page":   cast.ToInt(math.Ceil(cast.ToFloat64(totalNum) / cast.ToFloat64(reqData.PageSize))),
		"list":         users,
	}).FPrint(w)

}

func Info(ctx *context.Context, w http.ResponseWriter, req *http.Request) {

	domains := config.Instance.Domains
	domains = array.Difference(domains, []string{config.Instance.Domain})
	domains = append([]string{config.Instance.Domain}, domains...)

	response.NewSuccessResponse(map[string]any{
		"account":  ctx.UserAccount,
		"name":     ctx.UserName,
		"is_admin": ctx.IsAdmin,
		"domains":  domains,
	}).FPrint(w)
}

func EditUser(ctx *context.Context, w http.ResponseWriter, req *http.Request) {

	if !ctx.IsAdmin {
		response.NewErrorResponse(response.NoAccessPrivileges, "No Access Privileges", "").FPrint(w)
		return
	}

	reqBytes, err := io.ReadAll(req.Body)
	if err != nil {
		log.Errorf("%+v", err)
	}
	var reqData userCreateRequest
	err = json.Unmarshal(reqBytes, &reqData)
	if err != nil {
		log.Errorf("%+v", err)
	}

	if reqData.Id == 0 && reqData.Account == "" {
		response.NewErrorResponse(response.ParamsError, "Params Error", "").FPrint(w)
		return
	}
	var user models.User
	if reqData.Id != 0 {
		_, err = db.Instance.Where("id=?", reqData.Id).Get(&user)
		if err != nil {
			log.Errorf("SQL Error: %+v", err)
		}
	} else {
		_, err = db.Instance.Where("account=?", reqData.Account).Get(&user)
		if err != nil {
			log.Errorf("SQL Error: %+v", err)
		}
	}
	if user.ID == 0 {
		response.NewErrorResponse(response.ParamsError, "User not found", "").FPrint(w)
		return
	}
	if reqData.Username != "" && reqData.Username != user.Name {
		user.Name = reqData.Username
	}

	if reqData.Disabled != user.Disabled {
		user.Disabled = reqData.Disabled
	}
	if reqData.Password != "" {
		user.Password = password.Encode(reqData.Password)
	}

	num, err := db.Instance.ID(user.ID).Cols("name", "password", "disabled").Update(&user)

	if err != nil {
		response.NewErrorResponse(response.ServerError, err.Error(), "").FPrint(w)
		return
	}
	if num == 0 {
		response.NewErrorResponse(response.ServerError, "No Data Update", "").FPrint(w)
		return
	}

	response.NewSuccessResponse(user).FPrint(w)
}
