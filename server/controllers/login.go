package controllers

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"pmail/dto"
	"pmail/dto/response"
	"pmail/i18n"
	"pmail/models"
	"pmail/mysql"
	"pmail/session"
)

type loginRequest struct {
	Account  string `json:"account"`
	Password string `json:"password"`
}

func Login(ctx *dto.Context, w http.ResponseWriter, req *http.Request) {

	reqBytes, err := io.ReadAll(req.Body)
	if err != nil {
		log.Errorf("%+v", err)
	}
	var retData loginRequest
	err = json.Unmarshal(reqBytes, &retData)
	if err != nil {
		log.Errorf("%+v", err)
	}

	var user models.User

	encodePwd := md5Encode(md5Encode(retData.Password+"pmail") + "pmail2023")

	err = mysql.Instance.Get(&user, mysql.WithContext(ctx, "select * from user where account =? and password =?"),
		retData.Account, encodePwd)
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

func md5Encode(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}
