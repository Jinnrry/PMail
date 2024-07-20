package email

import (
	"encoding/json"
	"github.com/Jinnrry/pmail/dto/response"
	"github.com/Jinnrry/pmail/services/detail"
	"github.com/Jinnrry/pmail/utils/context"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
)

type emailDetailRequest struct {
	ID int `json:"id"`
}

func EmailDetail(ctx *context.Context, w http.ResponseWriter, req *http.Request) {
	reqBytes, err := io.ReadAll(req.Body)
	if err != nil {
		log.WithContext(ctx).Errorf("%+v", err)
	}
	var retData emailDetailRequest
	err = json.Unmarshal(reqBytes, &retData)
	if err != nil {
		log.WithContext(ctx).Errorf("%+v", err)
	}

	if retData.ID <= 0 {
		response.NewErrorResponse(response.ParamsError, "ID错误", "").FPrint(w)
		return
	}

	email, err := detail.GetEmailDetail(ctx, retData.ID, true)
	if err != nil {
		response.NewErrorResponse(response.ServerError, err.Error(), "").FPrint(w)
		return
	}

	response.NewSuccessResponse(email).FPrint(w)

}
