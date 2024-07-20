package controllers

import (
	"github.com/Jinnrry/pmail/utils/context"
	"net/http"
)

type HandlerFunc func(*context.Context, http.ResponseWriter, *http.Request)
