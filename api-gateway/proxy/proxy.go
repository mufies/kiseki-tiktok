package proxy

import (
    "net/http"
    "net/http/httputil"
    "net/url"
)


func NewProxy(targetURL string) http.Handler {
	target, _ := url.Parse(targetURL)
	return httputil.NewSingleHostReverseProxy(target)
}
