package setup

import (
	"fmt"
	"strings"

	"github.com/Jinnrry/pmail/i18n"
	"github.com/Jinnrry/pmail/services/auth"
	"github.com/Jinnrry/pmail/utils/context"
	"github.com/Jinnrry/pmail/utils/errors"
	"github.com/Jinnrry/pmail/utils/ip"
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
			{Type: "A", Host: strings.ReplaceAll(configData.WebDomain, "."+configData.Domain, ""), Value: ip.GetIp(), TTL: 3600, Tips: i18n.GetText(ctx.Lang, "ip_taps")},
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
