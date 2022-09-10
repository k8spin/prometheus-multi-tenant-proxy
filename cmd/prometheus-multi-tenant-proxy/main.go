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
					Name:  "port",
					Usage: "Port to expose this prometheus proxy",
					Value: 9092,
				}, &cli.StringFlag{
					Name:  "prometheus-endpoint",
					Usage: "Prometheus server endpoint",
					Value: "http://localhost:9091",
				}, &cli.StringSliceFlag{
					Name:  "unprotected-endpoints",
					Usage: "Unprotected endpoints (mostly for live/readiness probes)",
					Value: cli.NewStringSlice("/-/healthy", "/-/ready"),
				}, &cli.StringFlag{
					Name:  "auth-config",
					Usage: "AuthN yaml configuration file path",
					Value: "authn.yaml",
				}, &cli.IntFlag{
					Name:  "reload-interval",
					Usage: "Interval time to reload the authn configuration file (minutes)",
					Value: 5,
				},
			},
		},
	}
	app.Run(os.Args)
}
