package email

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"pmail/dto/response"
	"pmail/services/group"
	"pmail/utils/context"
)

type moveRequest struct {
	GroupId int   `json:"group_id"`
	IDs     []int `json:"ids"`
}

func Move(ctx *context.Context, w http.ResponseWriter, req *http.Request) {
	reqBytes, err := io.ReadAll(req.Body)
	if err != nil {
		log.WithContext(ctx).Errorf("%+v", err)
	}
	var reqData moveRequest
	err = json.Unmarshal(reqBytes, &reqData)
	if err != nil {
		log.WithContext(ctx).Errorf("%+v", err)
	}

	if len(reqData.IDs) <= 0 {
		response.NewErrorResponse(response.ParamsError, "ID错误", "").FPrint(w)
		return
	}

	if !group.MoveMailToGroup(ctx, reqData.IDs, reqData.GroupId) {
		response.NewErrorResponse(response.ServerError, "Error", "").FPrint(w)
		return
	}
	response.NewSuccessResponse("success").FPrint(w)

}
