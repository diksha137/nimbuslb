package backend

import (
	"net/http/httputil"
	"net/url"
)

type Backend struct {
	URL   *url.URL
	Proxy *httputil.ReverseProxy
	Name  string
}

func New(name, rawURL string) (*Backend, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	return &Backend{
		Name:  name,
		URL:   u,
		Proxy: httputil.NewSingleHostReverseProxy(u),
	}, nil
}
