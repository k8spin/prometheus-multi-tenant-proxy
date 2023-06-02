package proxy

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	injector "github.com/prometheus-community/prom-label-proxy/injectproxy"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql/parser"
)

type ReversePrometheusRoundTripper struct {
	prometheusServerURL *url.URL
}

func (r *ReversePrometheusRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	log.Printf("[TO]\t%s %s %s\n", req.RemoteAddr, req.Method, req.URL)
	return http.DefaultTransport.RoundTrip(req)
}

func (r *ReversePrometheusRoundTripper) Director(req *http.Request) {
	if req.URL.Path == "/api/v1/query" || req.URL.Path == "/api/v1/query_range" {
		if err := r.modifyRequest(req, "query"); err != nil {
			log.Printf("[ERROR]\t%s\n", err)
		}
	}
	if req.URL.Path == "/api/v1/series" {
		if err := r.modifyRequest(req, "match[]"); err != nil {
			log.Printf("[ERROR]\t%s\n", err)
		}
	}

	req.Host = r.prometheusServerURL.Host
	req.URL.Scheme = r.prometheusServerURL.Scheme
	req.URL.Host = r.prometheusServerURL.Host

	req.Header.Set("X-Forwarded-Host", req.Host)
	req.Header.Del("Authorization")
	req.Header.Del("Token")
}

func (r *ReversePrometheusRoundTripper) modifyRequest(req *http.Request, prometheusFormParameter string) error {

	namespaces := req.Context().Value(Namespaces).([]string)
	if len(namespaces) == 0 {
		return nil
	}

	var matcher labels.Matcher
	if len(namespaces) == 1 {
		// If there is only one namespace, we can use the more efficient MatchEqual matcher.
		matcher = labels.Matcher{
			Name:  "namespace",
			Type:  labels.MatchEqual,
			Value: namespaces[0],
		}
	} else {
		// If there are multiple namespaces, we need to use the MatchRegexp matcher.
		matcher = labels.Matcher{
			Name:  "namespace",
			Type:  labels.MatchRegexp,
			Value: strings.Join(namespaces, "|"),
		}
	}

	e := injector.NewEnforcer(false, []*labels.Matcher{
		&matcher,
	}...)

	if err := req.ParseForm(); err != nil {
		return err
	}

	form := req.Form

	for key, values := range form {
		value := values[0]
		if key == prometheusFormParameter {
			expr, err := parser.ParseExpr(value)
			if err != nil {
				return err
			}
			if err := e.EnforceNode(expr); err != nil {
				return err
			}
			value = expr.String()
		}
		form.Set(key, value)
	}

	newFormData := form.Encode()

	if req.Method == "POST" {
		req.Body = ioutil.NopCloser(strings.NewReader(newFormData))
		req.ContentLength = int64(len(newFormData))

	} else if req.Method == "GET" {
		req.URL.RawQuery = newFormData
	}

	return nil
}
