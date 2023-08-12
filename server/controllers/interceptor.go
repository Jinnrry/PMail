package controllers

import (
	"net/http"
	"pmail/config"
)

func Interceptor(w http.ResponseWriter, r *http.Request) {
	URL := "https://" + config.Instance.WebDomain + r.URL.Path
	http.Redirect(w, r, URL, http.StatusMovedPermanently)
}
