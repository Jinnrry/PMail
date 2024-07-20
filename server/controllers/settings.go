package controllers

import (
	"encoding/json"
	"github.com/Jinnrry/pmail/db"
	"github.com/Jinnrry/pmail/dto/response"
	"github.com/Jinnrry/pmail/i18n"
	"github.com/Jinnrry/pmail/utils/context"
	"github.com/Jinnrry/pmail/utils/password"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
)

type modifyPasswordRequest struct {
	Password string `json:"password"`
}

func ModifyPassword(ctx *context.Context, w http.ResponseWriter, req *http.Request) {
	reqBytes, err := io.ReadAll(req.Body)
	if err != nil {
		log.Errorf("%+v", err)
	}
	var retData modifyPasswordRequest
	err = json.Unmarshal(reqBytes, &retData)
	if err != nil {
		log.Errorf("%+v", err)
	}

	if retData.Password != "" {
		encodePwd := password.Encode(retData.Password)

		_, err := db.Instance.Exec(db.WithContext(ctx, "update user set password = ? where id =?"), encodePwd, ctx.UserID)
		if err != nil {
			response.NewErrorResponse(response.ServerError, i18n.GetText(ctx.Lang, "unknowError"), "").FPrint(w)
			return
		}

	}

	response.NewSuccessResponse(i18n.GetText(ctx.Lang, "succ")).FPrint(w)
}
