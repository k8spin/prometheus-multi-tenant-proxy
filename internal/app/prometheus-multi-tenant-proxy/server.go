package proxy

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/k8spin/prometheus-multi-tenant-proxy/internal/pkg"
	"github.com/urfave/cli/v2"
)

// Serve serves
func Serve(c *cli.Context) error {
	prometheusServerURL, _ := url.Parse(c.String("prometheus-endpoint"))
	serveAt := fmt.Sprintf(":%d", c.Int("port"))
	authConfigLocation := c.String("auth-config")
	authConfig, _ := pkg.ParseConfig(&authConfigLocation)

	reverseProxy := httputil.NewSingleHostReverseProxy(prometheusServerURL)
	handler := createHandler(reverseProxy, prometheusServerURL)
	http.HandleFunc("/-/healthy", LogRequest(handler))
	http.HandleFunc("/-/ready", LogRequest(handler))
	http.HandleFunc("/", LogRequest(BasicAuth(handler, authConfig)))
	if err := http.ListenAndServe(serveAt, nil); err != nil {
		log.Fatalf("Prometheus multi tenant proxy can not start %v", err)
		return err
	}
	return nil
}

func createHandler(proxy *httputil.ReverseProxy, prometheusServerURL *url.URL) http.HandlerFunc {
	return ReversePrometheus(proxy, prometheusServerURL)
}
