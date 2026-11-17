package controllers

import (
	"net/http"

	"github.com/Jinnrry/pmail/config"
	"github.com/Jinnrry/pmail/dto/response"
	"github.com/Jinnrry/pmail/services/domain"
	"github.com/Jinnrry/pmail/utils/context"
)

// GetDNSRecords 返回所有域名需要配置的 DNS 记录，仅管理员可用
func GetDNSRecords(ctx *context.Context, w http.ResponseWriter, req *http.Request) {
	if !ctx.IsAdmin {
		response.NewErrorResponse(response.NoAccessPrivileges, "No Access Privileges", "").FPrint(w)
		return
	}

	records, err := domain.GetDNSSettings(ctx, config.Instance, config.Instance.Domains)
	if err != nil {
		response.NewErrorResponse(response.ServerError, err.Error(), "").FPrint(w)
		return
	}

	response.NewSuccessResponse(records).FPrint(w)
}
