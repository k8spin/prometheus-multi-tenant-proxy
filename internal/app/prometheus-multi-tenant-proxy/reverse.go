package proxy

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/angelbarrera92/prometheus-multi-tenant-proxy/pkg/injector"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/promql"
)

// ReversePrometheus a
func ReversePrometheus(reverseProxy *httputil.ReverseProxy, prometheusLabelProxyServerURL *url.URL) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		checkRequest(r, prometheusLabelProxyServerURL)
		reverseProxy.ServeHTTP(w, r)
		log.Printf("[PROXY]\t%+v\n", r.URL)
	}
}

func modifyRequest(r *http.Request, prometheusLabelProxyServerURL *url.URL) error {
	namespace := r.Context().Value(Namespace)
	expr, err := promql.ParseExpr(r.FormValue("query"))
	if err != nil {
		return err
	}

	err = injector.SetRecursive(expr, []*labels.Matcher{
		{
			Name:  "namespace",
			Type:  labels.MatchEqual,
			Value: namespace.(string),
		},
	})
	if err != nil {
		return err
	}
	q := r.URL.Query()
	q.Set("query", expr.String())
	r.URL.RawQuery = q.Encode()
	return nil
}

func checkRequest(r *http.Request, prometheusLabelProxyServerURL *url.URL) error {
	if r.URL.Path == "/api/v1/query" || r.URL.Path == "/api/v1/query_range" {
		if err := modifyRequest(r, prometheusLabelProxyServerURL); err != nil {
			return err
		}
	}
	r.Host = prometheusLabelProxyServerURL.Host
	r.URL.Scheme = prometheusLabelProxyServerURL.Scheme
	r.URL.Host = prometheusLabelProxyServerURL.Host
	r.Header.Set("X-Forwarded-Host", r.Host)
	return nil
}
