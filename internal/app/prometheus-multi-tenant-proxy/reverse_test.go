package proxy

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"testing"
)

var (
	promURL = "http://prom.real"
	base, _ = url.Parse(promURL)
)

func ctx(namespaces []string, labels map[string][]string) context.Context {
	if namespaces == nil {
		namespaces = []string{}
	}
	if labels == nil {
		labels = map[string][]string{}
	}
	c := context.WithValue(context.TODO(), Namespaces, namespaces)
	c = context.WithValue(c, Labels, labels)
	return c
}

func getRequest(url string, namespaces []string, labels map[string][]string) *http.Request {
	r, _ := http.NewRequest(http.MethodGet, url, nil)
	return r.WithContext(ctx(namespaces, labels))
}

func ns2qs(ns []string) string {
	switch len(ns) {
	case 0:
		return ""
	case 1:
		return fmt.Sprintf("namespace=\"%s\"", ns[0])
	}
	return fmt.Sprintf("namespace=~\"%s\"", strings.Join(ns, "|"))
}

func labels2qs(labels map[string][]string) string {
	if len(labels) == 0 {
		return ""
	}
	matchers := make([]string, 0, len(labels))
	for k, v := range labels {
		// join the v with "|"
		matchers = append(matchers, fmt.Sprintf("%s=~\"%s\"", k, strings.Join(v, "|")))
	}
	sort.Strings(matchers)
	return strings.Join(matchers, ",")
}

func TestReverse_Proxy(t *testing.T) {
	testCases := []struct {
		url string
	}{
		{"http://real.prom/"},
		{"https://real.prom"},
		{"http://real.prom/some/path"},
	}

	for _, tc := range testCases {
		t.Run(tc.url, func(t *testing.T) {
			u, _ := url.Parse(tc.url)
			tripper := ReversePrometheusRoundTripper{
				prometheusServerURL: u,
			}

			r := getRequest("http://prom.proxy/api/v1", nil, nil)
			tripper.Director(r)

			if r.URL.Scheme != u.Scheme {
				t.Errorf("Wrong scheme: %s", r.URL.Scheme)
			}
			if r.URL.Host != u.Host {
				t.Errorf("Wrong host: %s", r.URL.Host)
			}
			if r.URL.Path != u.Path+"/api/v1" {
				t.Errorf("Wrong path: %s", r.URL.Path)
			}
		})
	}
}

func TestReverse_ModifyGet(t *testing.T) {
	tripper := ReversePrometheusRoundTripper{
		prometheusServerURL: base,
	}

	testCases := []struct {
		query string
	}{
		{"query?query=label&time=1685685673.187"},
		{"query_range?start=2023-04-17T13:37:00.781Z&end=2023-04-17T13:43:00.781Z&step=60s&query=label"},
		{"series?start=2023-04-17T13:37:00.781Z&end=2023-04-17T13:43:00.781Z&match[]=up"},
	}

	for _, tc := range testCases {
		t.Run(strings.Split(tc.query, "?")[0], func(t *testing.T) {
			// test labels injection
			for _, labels := range []map[string][]string{{"foo": {"true"}}, {"bar": {"one"}, "buzz": {"two"}}} {
				r := getRequest(fmt.Sprintf("%s/api/v1/%s", promURL, tc.query), nil, labels)
				tripper.Director(r)

				parsed, _ := url.QueryUnescape(r.URL.RawQuery)
				qls := labels2qs(labels)

				if !strings.Contains(parsed, qls) {
					t.Errorf("labels not injected: %s (looking for %s)", parsed, qls)
				}
			}
			// test namespace injection
			for _, ns := range [][]string{{"ns1"}, {"ns1", "ns2"}} {
				r := getRequest(fmt.Sprintf("%s/api/v1/%s", promURL, tc.query), ns, nil)
				tripper.Director(r)

				parsed, _ := url.QueryUnescape(r.URL.RawQuery)
				qns := ns2qs(ns)

				if !strings.Contains(parsed, qns) {
					t.Errorf("namespace not injected: %s (looking for %s)", parsed, qns)
				}
			}
			// test both
			ns := []string{"some-ns"}
			labels := map[string][]string{"some": {"label"}}
			r := getRequest(fmt.Sprintf("%s/api/v1/%s", promURL, tc.query), ns, labels)
			tripper.Director(r)

			parsed, _ := url.QueryUnescape(r.URL.RawQuery)
			qns := ns2qs(ns)
			qls := labels2qs(labels)

			if !strings.Contains(parsed, qns) || !strings.Contains(parsed, qls) {
				t.Errorf("namespace+labels not injected: %s (looking for %s and %s)", parsed, qns, qls)
			}
		})
	}
}

func TestReverse_NoNs(t *testing.T) {
	// A request without namespaces nor labels
	// will end up with a query string containing an empty prometheus query.
	tripper := ReversePrometheusRoundTripper{
		prometheusServerURL: base,
	}
	path := "/api/v1/query"
	expectedQuery := "query="

	// If the context contains an empty namespaces slice, the request isn't touched.
	r := getRequest(fmt.Sprintf("%s/api/v1/query?query=foo", promURL), nil, nil)
	tripper.Director(r)

	if r.URL.Path != path {
		t.Errorf("Path should have been preserved: %v", r.URL.Path)
	}
	parsed, _ := url.QueryUnescape(r.URL.RawQuery)
	if parsed != expectedQuery {
		t.Errorf("Query should not have been preserved: %v", r.URL.RawQuery)
	}
}

func TestReverse_Untouched(t *testing.T) {
	tripper := ReversePrometheusRoundTripper{
		prometheusServerURL: base,
	}
	ns := []string{"ns1"}
	qns := ns2qs(ns)

	testCases := []struct {
		query string
	}{
		// Those endpoints are checked for authentication, but their query is not modified.
		{"queryexamplars"},
		{"format_query?query=foo/bar"},
		{"labels"},
		{"label/foo/values"},
		{"targets"},
		{"target/metadata"},
		{"metadata"},
		{"status/config"},
		{"status/flags"},
		{"status/runtimeinfo"},
		{"status/buildinfo"},
		{"status/tsdb"}, // and all below
		{"status/walreplay"},
		{"rules"},
		{"alerts"},
	}

	for _, tc := range testCases {
		t.Run(tc.query, func(t *testing.T) {
			r := getRequest(fmt.Sprintf("%s/api/v1/%s", promURL, tc.query), ns, nil)
			tripper.Director(r)

			parsed, _ := url.QueryUnescape(r.URL.RawQuery)
			if strings.Contains(parsed, qns) {
				t.Errorf("namespace injected but shouldn't: %s", parsed)
			}
		})
	}

}
