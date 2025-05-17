package config

import "net/http"

type AnnictAuthTransport struct {
	Token     string
	Transport http.RoundTripper
}

func (t *AnnictAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	transport := t.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}
	req.Header.Set("Authorization", "Bearer "+t.Token)
	// User-Agent を設定することが推奨されます
	req.Header.Set("User-Agent", "AnnictSlackBot/1.0 (github.com/monchh/annict-slack-bot)")
	return transport.RoundTrip(req)
}
