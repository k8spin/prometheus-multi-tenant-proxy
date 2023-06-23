package main

import (
	"os"

	proxy "github.com/k8spin/prometheus-multi-tenant-proxy/internal/app/prometheus-multi-tenant-proxy"
	"github.com/urfave/cli/v2"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	envPrefix := "PROM_PROXY_"
	app := cli.NewApp()
	app.Name = "Prometheus multi-tenant proxy"
	app.Usage = "Makes your Prometheus server multi tenant"
	app.Version = version
	app.Authors = []*cli.Author{
		{Name: "Angel Barrera", Email: "angel@k8spin.cloud"},
		{Name: "Pau Rosello", Email: "pau@k8spin.cloud"},
	}
	app.Commands = []*cli.Command{
		{
			Name:   "run",
			Usage:  "Runs the Prometheus multi-tenant proxy",
			Action: proxy.Serve,
			Flags: []cli.Flag{
				&cli.IntFlag{
					Name:    "port",
					Usage:   "Port to expose this prometheus proxy",
					Value:   9092,
					EnvVars: []string{envPrefix + "PORT"},
				}, &cli.StringFlag{
					Name:    "prometheus-endpoint",
					Usage:   "Prometheus server endpoint",
					Value:   "http://localhost:9091",
					EnvVars: []string{envPrefix + "ENDPOINT"},
				}, &cli.StringSliceFlag{
					Name:    "unprotected-endpoints",
					Usage:   "Unprotected endpoints (mostly for live/readiness probes)",
					Value:   cli.NewStringSlice("/-/healthy", "/-/ready"),
					EnvVars: []string{envPrefix + "UNPROTECTED_ENDPOINTS"},
				}, &cli.StringFlag{
					Name:    "auth-type",
					Usage:   "Auth mechanism: one of 'basic' or 'jwt'",
					Value:   "basic",
					EnvVars: []string{envPrefix + "AUTH_TYPE"},
				}, &cli.StringFlag{
					Name:    "auth-config",
					Usage:   "AuthN yaml configuration file path (basic auth) or jwks file path/url (jwt auth)",
					Value:   "authn.yaml",
					EnvVars: []string{envPrefix + "AUTH_CONFIG"},
				}, &cli.IntFlag{
					Name:    "reload-interval",
					Usage:   "Interval time to reload the configuration (minutes)",
					Value:   5,
					EnvVars: []string{envPrefix + "RELOAD_INTERVAL"},
				}, &cli.BoolFlag{
					Name:    "aws",
					Value:   false,
					Usage:   "If true, sign the request using AWS credentials",
					EnvVars: []string{envPrefix + "USE_AWS"},
				},
			},
		},
	}
	app.Run(os.Args)
}
