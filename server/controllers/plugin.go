package controllers

import (
	"github.com/Jinnrry/pmail/dto/response"
	"github.com/Jinnrry/pmail/hooks"
	"github.com/Jinnrry/pmail/utils/context"
	"io"
	"net/http"
	"strings"
)

func GetPluginList(ctx *context.Context, w http.ResponseWriter, req *http.Request) {
	ret := []string{}
	for s, _ := range hooks.HookList {
		ret = append(ret, s)
	}
	response.NewSuccessResponse(ret).FPrint(w)

}

func SettingsHtml(ctx *context.Context, w http.ResponseWriter, req *http.Request) {
	args := strings.Split(req.RequestURI, "/")
	if len(args) < 4 {
		response.NewErrorResponse(response.ParamsError, "404", "").FPrint(w)
		return
	}

	pluginName := args[4]
	if plugin, ok := hooks.HookList[pluginName]; ok {
		dt, err := io.ReadAll(req.Body)
		if err != nil {
			response.NewErrorResponse(response.ParamsError, err.Error(), "").FPrint(w)
			return
		}
		html := plugin.SettingsHtml(ctx,
			strings.Join(args[4:], "/"),
			string(dt),
		)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		w.Write([]byte(html))
		return

	}
	response.NewErrorResponse(response.ParamsError, "404", "")
}
