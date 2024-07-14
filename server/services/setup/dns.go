package setup

import (
	"fmt"
	"pmail/i18n"
	"pmail/services/auth"
	"pmail/utils/context"
	"pmail/utils/errors"
	"pmail/utils/ip"
)

type DNSItem struct {
	Type  string `json:"type"`
	Host  string `json:"host"`
	Value string `json:"value"`
	TTL   int    `json:"ttl"`
	Tips  string `json:"tips"`
}

func GetDNSSettings(ctx *context.Context) (map[string][]*DNSItem, error) {
	configData, err := ReadConfig()
	if err != nil {
		return nil, errors.Wrap(err)
	}

	ret := make(map[string][]*DNSItem)

	for _, domain := range configData.Domains {
		ret[domain] = []*DNSItem{
			{Type: "A", Host: "smtp", Value: ip.GetIp(), TTL: 3600, Tips: i18n.GetText(ctx.Lang, "ip_taps")},
			{Type: "A", Host: "pop", Value: ip.GetIp(), TTL: 3600, Tips: i18n.GetText(ctx.Lang, "ip_taps")},
			{Type: "A", Host: "@", Value: ip.GetIp(), TTL: 3600, Tips: i18n.GetText(ctx.Lang, "ip_taps")},
			{Type: "MX", Host: "@", Value: fmt.Sprintf("smtp.%s", domain), TTL: 3600},
			{Type: "TXT", Host: "@", Value: "v=spf1 a mx ~all", TTL: 3600},
			{Type: "TXT", Host: "default._domainkey", Value: auth.DkimGen(), TTL: 3600},
		}
	}

	return ret, nil
}
