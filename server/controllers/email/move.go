package email

import (
	"encoding/json"
	"github.com/Jinnrry/pmail/dto/response"
	"github.com/Jinnrry/pmail/models"
	"github.com/Jinnrry/pmail/services/group"
	"github.com/Jinnrry/pmail/utils/context"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
)

type moveRequest struct {
	GroupId   int    `json:"group_id"`
	GroupName string `json:"group_name"`
	IDs       []int  `json:"ids"`
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

	if name, ok := models.GroupCodeToName[reqData.GroupId]; ok {
		err := group.Move2DefaultBox(ctx, reqData.IDs, name)
		if err != nil {
			response.NewErrorResponse(response.ServerError, "Error", err.Error()).FPrint(w)
			return
		}
	} else if !group.MoveMailToGroup(ctx, reqData.IDs, reqData.GroupId) {
		response.NewErrorResponse(response.ServerError, "Error", "").FPrint(w)
		return
	}
	response.NewSuccessResponse("success").FPrint(w)

}
