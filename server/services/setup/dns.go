package setup

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"pmail/i18n"
	"pmail/services/auth"
	"pmail/utils/context"
	"pmail/utils/errors"
)

type DNSItem struct {
	Type  string `json:"type"`
	Host  string `json:"host"`
	Value string `json:"value"`
	TTL   int    `json:"ttl"`
	Tips  string `json:"tips"`
}

func GetDNSSettings(ctx *context.Context) ([]*DNSItem, error) {
	configData, err := ReadConfig()
	if err != nil {
		return nil, errors.Wrap(err)
	}

	ret := []*DNSItem{
		{Type: "A", Host: "smtp", Value: getIp(), TTL: 3600, Tips: i18n.GetText(ctx.Lang, "ip_taps")},
		{Type: "MX", Host: "-", Value: fmt.Sprintf("smtp.%s", configData.Domain), TTL: 3600},
		{Type: "TXT", Host: "-", Value: "v=spf1 a mx ~all", TTL: 3600},
		{Type: "TXT", Host: "default._domainkey", Value: auth.DkimGen(), TTL: 3600},
	}
	return ret, nil
}

func getIp() string {
	resp, err := http.Get("http://ip-api.com/json/?lang=zh-CN ")
	if err != nil {
		return "Your Server IP"
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		body, err := io.ReadAll(resp.Body)
		if err == nil {
			var queryRes map[string]string
			_ = json.Unmarshal(body, &queryRes)

			return queryRes["query"]
		}
	}
	return "Your Server IP"
}
