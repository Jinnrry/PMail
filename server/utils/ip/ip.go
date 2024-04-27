package ip

import (
	"encoding/json"
	"io"
	"net/http"
)

var ip string

func GetIp() string {
	if ip != "" {
		return ip
	}

	resp, err := http.Get("http://ip-api.com/json/?lang=zh-CN ")
	if err != nil {
		return "[Your Server IP]"
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		body, err := io.ReadAll(resp.Body)
		if err == nil {
			var queryRes map[string]string
			_ = json.Unmarshal(body, &queryRes)
			ip = queryRes["query"]
			return queryRes["query"]
		}
	}
	return "[Your Server IP]"
}
