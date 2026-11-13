package controllers

import (
	"mime"
	"net/http"
	"strings"

	"github.com/Jinnrry/pmail/dto/response"
	"github.com/Jinnrry/pmail/services/attachments"
	"github.com/Jinnrry/pmail/utils/context"
	"github.com/spf13/cast"
)

func GetAttachments(ctx *context.Context, w http.ResponseWriter, req *http.Request) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Security-Policy", "sandbox")
	urlInfos := strings.Split(req.RequestURI, "/")
	if len(urlInfos) != 4 {
		response.NewErrorResponse(response.ParamsError, "", "").FPrint(w)
		return
	}
	emailId := cast.ToInt(urlInfos[2])
	cid := urlInfos[3]

	att := attachments.GetAttachments(ctx, emailId, cid)

	if att == nil || len(att.Content) == 0 {
		response.NewErrorResponse(response.ParamsError, "", "").FPrint(w)
		return
	}
	contentType, inline := attachments.InlineContentType(att.ContentType)
	w.Header().Set("Content-Type", contentType)
	if !inline {
		w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": att.Filename}))
	}
	w.Write(att.Content)
}

func Download(ctx *context.Context, w http.ResponseWriter, req *http.Request) {
	urlInfos := strings.Split(req.RequestURI, "/")
	if len(urlInfos) != 5 {
		response.NewErrorResponse(response.ParamsError, "", "").FPrint(w)
		return
	}
	emailId := cast.ToInt(urlInfos[3])
	index := cast.ToInt(urlInfos[4])

	fileName, content := attachments.GetAttachmentsByIndex(ctx, emailId, index)

	if len(content) == 0 {
		response.NewErrorResponse(response.ParamsError, "", "").FPrint(w)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": fileName}))
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Write(content)
}
