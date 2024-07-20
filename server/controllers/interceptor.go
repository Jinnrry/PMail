package controllers

import (
	"github.com/Jinnrry/pmail/config"
	"net/http"
)

func Interceptor(w http.ResponseWriter, r *http.Request) {
	URL := "https://" + config.Instance.WebDomain + r.URL.Path
	http.Redirect(w, r, URL, http.StatusMovedPermanently)
}
