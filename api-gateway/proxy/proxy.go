package proxy

import (
    "net/http"
    "net/http/httputil"
    "net/url"
)

func NewProxy(targetURL string) http.Handler {
	target, _ := url.Parse(targetURL)
	proxy := httputil.NewSingleHostReverseProxy(target)

	// Wrap the proxy to remove CORS headers from backend responses
	// The gateway's CORS middleware will set these instead
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
	}

	originalModifyResponse := proxy.ModifyResponse
	proxy.ModifyResponse = func(resp *http.Response) error {
		// Remove CORS headers from backend service responses
		// to prevent duplicate headers (gateway sets these)
		resp.Header.Del("Access-Control-Allow-Origin")
		resp.Header.Del("Access-Control-Allow-Credentials")
		resp.Header.Del("Access-Control-Allow-Headers")
		resp.Header.Del("Access-Control-Allow-Methods")
		resp.Header.Del("Access-Control-Expose-Headers")

		if originalModifyResponse != nil {
			return originalModifyResponse(resp)
		}
		return nil
	}

	return proxy
}
