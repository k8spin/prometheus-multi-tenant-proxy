package proxy

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/urfave/cli/v2"
)

// Serve serves
func Serve(c *cli.Context) error {
	prometheusServerURL, _ := url.Parse(c.String("prometheus-endpoint"))
	serveAt := fmt.Sprintf(":%d", c.Int("port"))
	authConfigLocation := c.String("auth-config")
	reloadInterval := c.Int("reload-interval")
	authType := c.String("auth-type")
	awsSign := c.Bool("aws")

	var auth Auth
	if authType == "basic" {
		auth = NewBasicAuth(authConfigLocation)
	} else if authType == "jwt" {
		auth = NewJwtAuth(authConfigLocation)
	} else {
		log.Fatalf("auth-type must be one of: basic, jwt") // will exit
	}

	if reloadInterval > 0 {
		ticker := time.NewTicker(time.Duration(reloadInterval) * time.Minute)
		quit := make(chan struct{})
		go func() {
			for {
				select {
				case <-ticker.C:
					auth.Load()
				case <-quit:
					ticker.Stop()
					return
				}
			}
		}()
	}

	rprt := ReversePrometheusRoundTripper{
		prometheusServerURL: prometheusServerURL,
	}

	director := rprt.Director

	if awsSign {
		signer := NewAWSSigner()
		director = signer.SignAfter(director)
		log.Printf("AWS signature enabled: %v", signer)
	}

	reverseProxy := httputil.ReverseProxy{
		Director:  director,
		Transport: &rprt,
	}

	for _, selected := range c.StringSlice("unprotected-endpoints") {
		log.Printf("Serving as unprotected endpoint: %s", selected)
		http.HandleFunc(selected, LogRequest(reverseProxy.ServeHTTP))
	}

	http.HandleFunc("/", LogRequest(AuthHandler(auth, reverseProxy.ServeHTTP)))
	if err := http.ListenAndServe(serveAt, nil); err != nil {
		log.Fatalf("Prometheus multi tenant proxy can not start %v", err)
		return err
	}
	return nil
}
