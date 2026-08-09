package proxy

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func NewReverseProxy(target string) (*httputil.ReverseProxy, error) {
	targetURL, err := url.Parse(target)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf(
			"Proxy error: method=%s path=%s error=%v",
			r.Method,
			r.URL.Path,
			err,
		)

		http.Error(
			w,
			"Bad Gateway",
			http.StatusBadGateway,
		)
	}

	return proxy, nil
}
