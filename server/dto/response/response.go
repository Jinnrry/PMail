package response

import (
	"encoding/json"
	"net/http"
)

const (
	NeedSetup          = 402
	NeedLogin          = 403
	NoAccessPrivileges = 405
	ParamsError        = 100
	ServerError        = 500
)

type Response struct {
	ErrorNo  int    `json:"errorNo"`
	ErrorMsg string `json:"errorMsg"`
	Data     any    `json:"data"`
}

func (p *Response) FPrint(w http.ResponseWriter) {
	bytesData, _ := json.Marshal(p)
	w.Write(bytesData)
}

func NewSuccessResponse(data any) *Response {
	return &Response{
		Data: data,
	}
}

func NewErrorResponse(errorNo int, errorMsg string, data any) *Response {
	return &Response{
		ErrorNo:  errorNo,
		ErrorMsg: errorMsg,
		Data:     data,
	}
}
