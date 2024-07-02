package controllers

import (
	"net/http"
	"pmail/dto/response"
)

func Ping(w http.ResponseWriter, req *http.Request) {
	response.NewSuccessResponse("pong").FPrint(w)
}
