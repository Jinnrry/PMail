package controllers

import (
	"database/sql"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"pmail/db"
	"pmail/dto/response"
	"pmail/i18n"
	"pmail/models"
	"pmail/session"
	"pmail/utils/context"
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
	
	_, err = db.Instance.Where("account =? and password =?", reqData.Account, encodePwd).Get(&user)
	if err != nil && err != sql.ErrNoRows {
		log.Errorf("%+v", err)
	}

	if user.ID != 0 {
		userStr, _ := json.Marshal(user)
		session.Instance.Put(req.Context(), "user", string(userStr))
		response.NewSuccessResponse("").FPrint(w)
	} else {
		response.NewErrorResponse(response.ParamsError, i18n.GetText(ctx.Lang, "aperror"), "").FPrint(w)
	}
}
