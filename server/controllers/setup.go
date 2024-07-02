package controllers

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"pmail/config"
	"pmail/dto/response"
	"pmail/services/setup"
	"pmail/services/setup/ssl"
	"pmail/utils/context"
	"strings"
)

func AcmeChallenge(w http.ResponseWriter, r *http.Request) {
	log.Infof("AcmeChallenge: %s", r.URL.Path)
	instance := ssl.GetHttpChallengeInstance()
	token := strings.ReplaceAll(r.URL.Path, "/.well-known/acme-challenge/", "")
	auth, exist := instance.AuthInfo[token]
	if exist {
		w.Write([]byte(auth.KeyAuth))
	} else {
		log.Errorf("AcmeChallenge Error Token Infos:%+v", instance.AuthInfo)
		http.NotFound(w, r)
	}
}

type sslResponse struct {
	Port int    `json:"port"`
	Type string `json:"type"`
}

func Setup(ctx *context.Context, w http.ResponseWriter, req *http.Request) {
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
		dbType, dbDSN, err := setup.GetDatabaseSettings(ctx)
		if err != nil {
			response.NewErrorResponse(response.ServerError, err.Error(), "").FPrint(w)
			return
		}

		response.NewSuccessResponse(map[string]string{
			"db_type": dbType,
			"db_dsn":  dbDSN,
		}).FPrint(w)
		return
	}

	if reqData["step"] == "database" && reqData["action"] == "set" {
		err := setup.SetDatabaseSettings(ctx, reqData["db_type"], reqData["db_dsn"])
		if err != nil {
			response.NewErrorResponse(response.ServerError, err.Error(), "").FPrint(w)
			return
		}

		response.NewSuccessResponse("Succ").FPrint(w)
		return
	}

	if reqData["step"] == "password" && reqData["action"] == "get" {
		ok, err := setup.GetAdminPassword(ctx)
		if err != nil {
			response.NewErrorResponse(response.ServerError, err.Error(), "").FPrint(w)
			return
		}
		response.NewSuccessResponse(ok).FPrint(w)
		return
	}

	if reqData["step"] == "password" && reqData["action"] == "set" {
		err := setup.SetAdminPassword(ctx, reqData["account"], reqData["password"])
		if err != nil {
			response.NewErrorResponse(response.ServerError, err.Error(), "").FPrint(w)
			return
		}
		response.NewSuccessResponse("Succ").FPrint(w)
		return
	}

	if reqData["step"] == "domain" && reqData["action"] == "get" {
		smtpDomain, webDomain, err := setup.GetDomainSettings()
		if err != nil {
			response.NewErrorResponse(response.ServerError, err.Error(), "").FPrint(w)
			return
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
			response.NewErrorResponse(response.ServerError, err.Error(), "").FPrint(w)
			return
		}
		response.NewSuccessResponse("Succ").FPrint(w)
		return
	}

	if reqData["step"] == "dns" && reqData["action"] == "get" {
		dnsInfos, err := setup.GetDNSSettings(ctx)
		if err != nil {
			response.NewErrorResponse(response.ServerError, err.Error(), "").FPrint(w)
			return
		}
		response.NewSuccessResponse(dnsInfos).FPrint(w)
		return
	}

	if reqData["step"] == "ssl" && reqData["action"] == "get" {
		sslType := ssl.GetSSL()
		res := sslResponse{
			Type: sslType,
			Port: config.Instance.GetSetupPort(),
		}
		response.NewSuccessResponse(res).FPrint(w)
		return
	}

	if reqData["step"] == "ssl" && reqData["action"] == "getParams" {
		params, err := ssl.GetServerParamsList(reqData["serverName"])
		if err != nil {
			response.NewErrorResponse(response.ServerError, err.Error(), "").FPrint(w)
			return
		}
		response.NewSuccessResponse(params).FPrint(w)
		return
	}

	if reqData["step"] == "ssl" && reqData["action"] == "setParams" {
		for key, v := range reqData {
			if key != "step" && key != "action" {
				ssl.SetDomainServerParams(key, v)
			}
		}
		response.NewSuccessResponse("Succ").FPrint(w)
		return
	}

	if reqData["step"] == "ssl" && reqData["action"] == "set" {

		serviceName, ok := reqData["serviceName"]
		if !ok {
			serviceName = ""
		}
		err := ssl.SetSSL(reqData["ssl_type"], reqData["key_path"], reqData["crt_path"], serviceName)
		if err != nil {
			response.NewErrorResponse(response.ServerError, err.Error(), "").FPrint(w)
			return
		}

		if reqData["ssl_type"] == config.SSLTypeAutoHTTP || reqData["ssl_type"] == config.SSLTypeAutoDNS {
			err = ssl.GenSSL(false)
			if err != nil {
				response.NewErrorResponse(response.ServerError, err.Error(), "").FPrint(w)
				return
			}
		}

		response.NewSuccessResponse("Succ").FPrint(w)

		if reqData["ssl_type"] == config.SSLTypeUser {
			setup.Finish()
		}
		return
	}

}
