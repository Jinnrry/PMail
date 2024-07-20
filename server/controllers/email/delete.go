package email

import (
	"encoding/json"
	"github.com/Jinnrry/pmail/dto/response"
	"github.com/Jinnrry/pmail/services/del_email"
	"github.com/Jinnrry/pmail/utils/context"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
)

type emailDeleteRequest struct {
	IDs       []int64 `json:"ids"`
	ForcedDel bool    `json:"forcedDel"`
}

func EmailDelete(ctx *context.Context, w http.ResponseWriter, req *http.Request) {
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

	err = del_email.DelEmail(ctx, reqData.IDs, reqData.ForcedDel)
	if err != nil {
		response.NewErrorResponse(response.ServerError, err.Error(), "").FPrint(w)
		return
	}
	response.NewSuccessResponse("success").FPrint(w)

}
