package ssl

import (
	"github.com/Jinnrry/pmail/utils/context"
	"github.com/go-acme/lego/v4/challenge/dns01"
	log "github.com/sirupsen/logrus"
	"time"
)

type authInfo struct {
	Domain  string
	Token   string
	KeyAuth string
}

type HttpChallenge struct {
	AuthInfo map[string]*authInfo
}

var instance *HttpChallenge

func (h *HttpChallenge) Present(domain, token, keyAuth string) error {
	h.AuthInfo[token] = &authInfo{
		Domain:  domain,
		Token:   token,
		KeyAuth: keyAuth,
	}

	return nil
}

func (h *HttpChallenge) CleanUp(domain, token, keyAuth string) error {
	delete(h.AuthInfo, token)
	return nil
}

func GetHttpChallengeInstance() *HttpChallenge {
	if instance == nil {
		instance = &HttpChallenge{
			AuthInfo: map[string]*authInfo{},
		}
	}
	return instance
}

type DNSChallenge struct {
	AuthInfo map[string]*authInfo
}

var dnsInstance *DNSChallenge

func GetDnsChallengeInstance() *DNSChallenge {
	if dnsInstance == nil {
		dnsInstance = &DNSChallenge{
			AuthInfo: map[string]*authInfo{},
		}
	}
	return dnsInstance
}

func (h *DNSChallenge) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)
	log.Infof("Presenting challenge Info : %+v", info)
	h.AuthInfo[token] = &authInfo{
		Domain:  info.FQDN,
		Token:   token,
		KeyAuth: info.Value,
	}
	log.Infof("SSL Log:%s %s %s", domain, token, keyAuth)
	return nil
}

func (h *DNSChallenge) CleanUp(domain, token, keyAuth string) error {
	delete(h.AuthInfo, token)
	return nil
}

func (h *DNSChallenge) Timeout() (timeout, interval time.Duration) {
	return 60 * time.Minute, 5 * time.Second
}

type DNSItem struct {
	Type  string `json:"type"`
	Host  string `json:"host"`
	Value string `json:"value"`
	TTL   int    `json:"ttl"`
	Tips  string `json:"tips"`
}

func (h *DNSChallenge) GetDNSSettings(ctx *context.Context) []*DNSItem {
	ret := []*DNSItem{}
	for _, info := range h.AuthInfo {
		ret = append(ret, &DNSItem{
			Type:  "TXT",
			Host:  info.Domain,
			Value: info.KeyAuth,
			TTL:   3600,
		})
	}

	return ret
}
