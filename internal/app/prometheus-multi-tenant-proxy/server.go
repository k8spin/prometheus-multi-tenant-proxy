package proxy

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/k8spin/prometheus-multi-tenant-proxy/internal/pkg"
	"github.com/urfave/cli/v2"
)

var (
	config     *pkg.Authn
	configLock = new(sync.RWMutex)
)

// Serve serves
func Serve(c *cli.Context) error {
	prometheusServerURL, _ := url.Parse(c.String("prometheus-endpoint"))
	serveAt := fmt.Sprintf(":%d", c.Int("port"))
	authConfigLocation := c.String("auth-config")
	reloadInterval := c.Int("reload-interval")

	loadConfig(authConfigLocation)
	ticker := time.NewTicker(time.Duration(reloadInterval) * time.Minute)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				loadConfig(authConfigLocation)
				log.Printf("Reloaded config file %s", authConfigLocation)
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	rprt := ReversePrometheusRoundTripper{
		prometheusServerURL: prometheusServerURL,
	}

	reverseProxy := httputil.ReverseProxy{
		Director:  rprt.Director,
		Transport: &rprt,
	}

	http.HandleFunc("/-/healthy", LogRequest(reverseProxy.ServeHTTP))
	http.HandleFunc("/-/ready", LogRequest(reverseProxy.ServeHTTP))
	http.HandleFunc("/", LogRequest(BasicAuth(reverseProxy.ServeHTTP)))
	if err := http.ListenAndServe(serveAt, nil); err != nil {
		log.Fatalf("Prometheus multi tenant proxy can not start %v", err)
		return err
	}
	return nil
}

func loadConfig(location string) {
	temp, err := pkg.ParseConfig(&location)
	if err != nil {
		log.Fatalf("Could not parse config file %s: %v", location, err)
		os.Exit(1)
	}
	configLock.Lock()
	config = temp
	configLock.Unlock()
}

func GetConfig() *pkg.Authn {
	configLock.RLock()
	defer configLock.RUnlock()
	return config
}
