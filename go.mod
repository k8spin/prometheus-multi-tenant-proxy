module github.com/k8spin/prometheus-multi-tenant-proxy

go 1.15

replace github.com/openshift/prom-label-proxy => github.com/prometheus-community/prom-label-proxy v0.2.1-0.20210129135803-4c30ca94e827

require (
	github.com/openshift/prom-label-proxy v0.1.1-0.20201207234304-88d4df554125
	github.com/prometheus/prometheus/v2/v2 v2.25.0
	github.com/urfave/cli/v2 v2.3.0
	gopkg.in/yaml.v2 v2.4.0
)
