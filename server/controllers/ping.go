package controllers

import (
	log "github.com/sirupsen/logrus"
	"net/http"
	"pmail/dto"
	"pmail/dto/response"
)

func Ping(ctx *dto.Context, w http.ResponseWriter, req *http.Request) {
	response.NewSuccessResponse("pong").FPrint(w)
	log.WithContext(ctx).Info("pong")
}
