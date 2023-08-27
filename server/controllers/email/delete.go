package email

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"pmail/dto"
	"pmail/dto/response"
	"pmail/services/del_email"
)

type emailDeleteRequest struct {
	IDs []int `json:"ids"`
}

func EmailDelete(ctx *dto.Context, w http.ResponseWriter, req *http.Request) {
	reqBytes, err := io.ReadAll(req.Body)
	if err != nil {
		log.WithContext(ctx).Errorf("%+v", err)
	}
	var reqData emailDeleteRequest
	err = json.Unmarshal(reqBytes, &reqData)
	if err != nil {
		log.WithContext(ctx).Errorf("%+v", err)
	}

	if len(reqData.IDs) <= 0 {
		response.NewErrorResponse(response.ParamsError, "ID错误", "").FPrint(w)
		return
	}

	err = del_email.DelEmail(ctx, reqData.IDs)
	if err != nil {
		response.NewErrorResponse(response.ServerError, err.Error(), "").FPrint(w)
		return
	}
	response.NewSuccessResponse("success").FPrint(w)

}
