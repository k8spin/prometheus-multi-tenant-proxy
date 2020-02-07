package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"log"
)

// ReversePrometheus a
func ReversePrometheus(reverseProxy *httputil.ReverseProxy, prometheusLabelProxyServerURL *url.URL) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		modifyRequest(r, prometheusLabelProxyServerURL)
		reverseProxy.ServeHTTP(w, r)
		log.Printf("[PROXY]\t%+v\n", r.URL)
	}
}

func modifyRequest(r *http.Request, prometheusLabelProxyServerURL *url.URL) {
	r.URL.Scheme = prometheusLabelProxyServerURL.Scheme
	r.URL.Host = prometheusLabelProxyServerURL.Host
	r.Host = prometheusLabelProxyServerURL.Host
	namespace := r.Context().Value(Namespace)
	r.Header.Set("X-Forwarded-Host", r.Host)
	q := r.URL.Query()
	q.Add("namespace", namespace.(string))
	r.URL.RawQuery = q.Encode()
}
