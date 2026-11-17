package domain

import (
	"strings"

	"github.com/Jinnrry/pmail/config"
	"github.com/Jinnrry/pmail/dto/parsemail"
	"github.com/Jinnrry/pmail/i18n"
	"github.com/Jinnrry/pmail/utils/context"
	"github.com/Jinnrry/pmail/utils/errors"
	"github.com/Jinnrry/pmail/utils/ip"
)

// 需要添加到域名服务商的 DNS 记录结构
type DNSItem struct {
	Type  string `json:"type"`
	Host  string `json:"host"`
	Value string `json:"value"`
	TTL   int    `json:"ttl"`
	Tips  string `json:"tips"`
}

// 返回 domains 中每个域名需要配置的 DNS 记录
func GetDNSSettings(ctx *context.Context, cfg *config.Config, domains []string) (map[string][]*DNSItem, error) {
	dkimRecord, err := parsemail.DkimRecord(cfg.DkimPrivateKeyPath)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	serverIP := ip.GetIp()
	ipTips := i18n.GetText(ctx.Lang, "ip_taps")

	ret := make(map[string][]*DNSItem, len(domains))
	for _, d := range domains {
		var records []*DNSItem
		if host, ok := strings.CutSuffix(cfg.WebDomain, "."+d); ok {
			records = append(records, &DNSItem{Type: "A", Host: host, Value: serverIP, TTL: 3600, Tips: ipTips})
		}
		ret[d] = append(records,
			&DNSItem{Type: "A", Host: "smtp", Value: serverIP, TTL: 3600, Tips: ipTips},
			&DNSItem{Type: "A", Host: "imap", Value: serverIP, TTL: 3600, Tips: ipTips},
			&DNSItem{Type: "A", Host: "pop", Value: serverIP, TTL: 3600, Tips: ipTips},
			&DNSItem{Type: "A", Host: "@", Value: serverIP, TTL: 3600, Tips: ipTips},
			&DNSItem{Type: "MX", Host: "@", Value: "smtp." + d, TTL: 3600},
			&DNSItem{Type: "TXT", Host: "@", Value: "v=spf1 a mx ~all", TTL: 3600},
			&DNSItem{Type: "TXT", Host: "default._domainkey", Value: dkimRecord, TTL: 3600},
		)
	}
	return ret, nil
}
