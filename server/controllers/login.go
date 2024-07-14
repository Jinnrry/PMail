package controllers

import (
	"database/sql"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"pmail/config"
	"pmail/db"
	"pmail/dto/response"
	"pmail/i18n"
	"pmail/models"
	"pmail/session"
	"pmail/utils/array"
	"pmail/utils/context"
	"pmail/utils/errors"
	"pmail/utils/password"
)

type loginRequest struct {
	Account  string `json:"account"`
	Password string `json:"password"`
}

func Login(ctx *context.Context, w http.ResponseWriter, req *http.Request) {

	reqBytes, err := io.ReadAll(req.Body)
	if err != nil {
		log.Errorf("%+v", err)
	}
	var reqData loginRequest
	err = json.Unmarshal(reqBytes, &reqData)
	if err != nil {
		log.Errorf("%+v", err)
	}

	var user models.User

	encodePwd := password.Encode(reqData.Password)
	_, err = db.Instance.Where("account =? and password =? and disabled=0", reqData.Account, encodePwd).Get(&user)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Errorf("%+v", err)
	}

	if user.ID != 0 {
		userStr, _ := json.Marshal(user)
		session.Instance.Put(req.Context(), "user", string(userStr))

		domains := config.Instance.Domains
		domains = array.Difference(domains, []string{config.Instance.Domain})
		domains = append([]string{config.Instance.Domain}, domains...)

		response.NewSuccessResponse(map[string]any{
			"account":  user.Account,
			"name":     user.Name,
			"is_admin": user.IsAdmin,
			"domains":  domains,
		}).FPrint(w)
	} else {
		response.NewErrorResponse(response.ParamsError, i18n.GetText(ctx.Lang, "aperror"), "").FPrint(w)
	}
}

func Logout(ctx *context.Context, w http.ResponseWriter, req *http.Request) {
	session.Instance.Clear(ctx.Context)
	response.NewSuccessResponse("Success").FPrint(w)
}
