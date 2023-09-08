package controllers

import (
	log "github.com/sirupsen/logrus"
	"net/http"
	"pmail/dto/response"
	"pmail/utils/context"
)

func Ping(ctx *context.Context, w http.ResponseWriter, req *http.Request) {
	response.NewSuccessResponse("pong").FPrint(w)
	log.WithContext(ctx).Info("pong")
}
