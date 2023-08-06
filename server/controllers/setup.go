package controllers

import (
	"encoding/json"
	"io"
	"net/http"
	"pmail/dto"
	"pmail/dto/response"
	"pmail/services/setup"
)

func Proxy(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("proxy"))
}

func Setup(ctx *dto.Context, w http.ResponseWriter, req *http.Request) {
	reqBytes, err := io.ReadAll(req.Body)
	if err != nil {
		response.NewSuccessResponse("").FPrint(w)
		return
	}

	var reqData map[string]string
	err = json.Unmarshal(reqBytes, &reqData)

	if err != nil {
		response.NewSuccessResponse("").FPrint(w)
		return
	}

	if reqData["step"] == "database" && reqData["action"] == "get" {
		dbType, dbDSN, err := setup.GetDatabaseSettings()
		if err != nil {
			response.NewErrorResponse(response.ServerError, err.Error(), "")
		}

		response.NewSuccessResponse(map[string]string{
			"db_type": dbType,
			"db_dsn":  dbDSN,
		}).FPrint(w)
		return
	}

	if reqData["step"] == "database" && reqData["action"] == "set" {
		err := setup.SetDatabaseSettings(reqData["db_type"], reqData["db_dsn"])
		if err != nil {
			response.NewErrorResponse(response.ServerError, err.Error(), "")
		}

		response.NewSuccessResponse("Succ").FPrint(w)
		return
	}

	if reqData["step"] == "domain" && reqData["action"] == "get" {
		smtpDomain, webDomain, err := setup.GetDomainSettings()
		if err != nil {
			response.NewErrorResponse(response.ServerError, err.Error(), "")
		}
		response.NewSuccessResponse(map[string]string{
			"smtp_domain": smtpDomain,
			"web_domain":  webDomain,
		}).FPrint(w)
		return
	}

	if reqData["step"] == "domain" && reqData["action"] == "set" {
		err := setup.SetDomainSettings(reqData["smtp_domain"], reqData["web_domain"])
		if err != nil {
			response.NewErrorResponse(response.ServerError, err.Error(), "")
		}
		response.NewSuccessResponse("Succ").FPrint(w)
		return
	}

	if reqData["step"] == "dns" && reqData["action"] == "get" {
		dnsInfos, err := setup.GetDNSSettings(ctx)
		if err != nil {
			response.NewErrorResponse(response.ServerError, err.Error(), "")
		}
		response.NewSuccessResponse(dnsInfos).FPrint(w)
		return
	}

	if reqData["step"] == "ssl" && reqData["action"] == "get" {
		err := setup.GenSSL()
		if err != nil {
			response.NewErrorResponse(response.ServerError, err.Error(), "")
		}
		response.NewSuccessResponse("").FPrint(w)
		return
	}
}
