package controllers

import (
	"encoding/json"
	"github.com/Jinnrry/pmail/config"
	"github.com/Jinnrry/pmail/dto/response"
	"github.com/Jinnrry/pmail/services/setup"
	"github.com/Jinnrry/pmail/services/setup/ssl"
	"github.com/Jinnrry/pmail/utils/context"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"io"
	"net/http"
	"os"
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
		response.NewErrorResponse(response.ServerError, "", err.Error()).FPrint(w)
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
		err := setup.SetDatabaseSettings(ctx, cast.ToString(reqData["db_type"]), cast.ToString(reqData["db_dsn"]))
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
		err := setup.SetAdminPassword(ctx, cast.ToString(reqData["account"]), cast.ToString(reqData["password"]))
		if err != nil {
			response.NewErrorResponse(response.ServerError, err.Error(), "").FPrint(w)
			return
		}
		response.NewSuccessResponse("Succ").FPrint(w)
		return
	}

	if reqData["step"] == "domain" && reqData["action"] == "get" {
		smtpDomain, webDomain, domains, smtpPort, imapPort, pop3Port, err := setup.GetDomainSettings()
		if err != nil {
			response.NewErrorResponse(response.ServerError, err.Error(), "").FPrint(w)
			return
		}
		response.NewSuccessResponse(map[string]any{
			"smtp_domain": smtpDomain,
			"web_domain":  webDomain,
			"domains":     domains,
			"smtp_port":   smtpPort,
			"imap_port":   imapPort,
			"pop3_port":   pop3Port,
		}).FPrint(w)
		return
	}

	if reqData["step"] == "domain" && reqData["action"] == "set" {
		err := setup.SetDomainSettings(
			cast.ToString(reqData["smtp_domain"]),
			cast.ToString(reqData["web_domain"]),
			reqData["multi_domain"],
			cast.ToInt(reqData["smtp_port"]),
			cast.ToInt(reqData["imap_port"]),
			cast.ToInt(reqData["pop3_port"]),
		)
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
		dnsChallenge := ssl.GetDnsChallengeInstance()

		response.NewSuccessResponse(dnsChallenge.GetDNSSettings(ctx)).FPrint(w)
		return
	}

	if reqData["step"] == "ssl" && reqData["action"] == "set" {

		if reqData["ssl_type"] == config.SSLTypeUser {
			keyPath := reqData["key_path"]
			crtPath := reqData["crt_path"]

			_, err := os.Stat(cast.ToString(keyPath))
			if err != nil {
				response.NewErrorResponse(response.ServerError, err.Error(), "").FPrint(w)
				return
			}

			_, err = os.Stat(cast.ToString(crtPath))
			if err != nil {
				response.NewErrorResponse(response.ServerError, err.Error(), "").FPrint(w)
				return
			}
		}

		err = ssl.SetSSL(cast.ToString(reqData["ssl_type"]), cast.ToString(reqData["key_path"]), cast.ToString(reqData["crt_path"]))
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
