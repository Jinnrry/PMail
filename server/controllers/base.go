package controllers

import (
	"net/http"
	"pmail/utils/context"
)

type HandlerFunc func(*context.Context, http.ResponseWriter, *http.Request)
