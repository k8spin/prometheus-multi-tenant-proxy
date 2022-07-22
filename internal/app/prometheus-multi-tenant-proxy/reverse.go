package proxy

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	injector "github.com/prometheus-community/prom-label-proxy/injectproxy"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql/parser"
)

// ReversePrometheus a
func ReversePrometheus(reverseProxy *httputil.ReverseProxy, prometheusServerURL *url.URL) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		checkRequest(r, prometheusServerURL)
		reverseProxy.ServeHTTP(w, r)
		log.Printf("[TO]\t%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
	}
}

func modifyRequest(r *http.Request, prometheusServerURL *url.URL, prometheusQueryParameter string) error {
	namespace := r.Context().Value(Namespace)
	expr, err := parser.ParseExpr(r.FormValue(prometheusQueryParameter))
	if err != nil {
		return err
	}

	e := injector.NewEnforcer(true, []*labels.Matcher{
		{
			Name:  "namespace",
			Type:  labels.MatchEqual,
			Value: namespace.(string),
		},
	}...)

	if err := e.EnforceNode(expr); err != nil {
		return err
	}

	q := r.URL.Query()
	q.Set(prometheusQueryParameter, expr.String())
	r.URL.RawQuery = q.Encode()
	return nil
}

func checkRequest(r *http.Request, prometheusServerURL *url.URL) error {
	if r.URL.Path == "/api/v1/query" || r.URL.Path == "/api/v1/query_range" {
		if err := modifyRequest(r, prometheusServerURL, "query"); err != nil {
			return err
		}
	}
	if r.URL.Path == "/api/v1/series" {
		if err := modifyRequest(r, prometheusServerURL, "match[]"); err != nil {
			return err
		}
	}
	r.Host = prometheusServerURL.Host
	r.URL.Scheme = prometheusServerURL.Scheme
	r.URL.Host = prometheusServerURL.Host
	r.Header.Set("X-Forwarded-Host", r.Host)
	return nil
}
