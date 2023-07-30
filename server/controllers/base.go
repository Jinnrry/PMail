package controllers

import (
	"net/http"
	"pmail/dto"
)

type HandlerFunc func(*dto.Context, http.ResponseWriter, *http.Request)
