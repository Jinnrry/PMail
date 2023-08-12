package ssl

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
