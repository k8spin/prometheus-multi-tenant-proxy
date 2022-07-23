# Prometheus — Multi-tenant proxy

![Build Status](https://action-badges.now.sh/k8spin/prometheus-multi-tenant-proxy)
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)

[![Twitter](https://img.shields.io/twitter/url/https/twitter.com/k8spin.svg?style=social&label=Follow%20%40k8spin)](https://twitter.com/k8spin)
[![Join the chat at https://slack.kubernetes.io](https://img.shields.io/badge/style-register-green.svg?style=social&label=Slack)](https://slack.kubernetes.io)

------

This project aims to make it easy to deploy a [Prometheus Server](https://github.com/prometheus/prometheus)
in a multi-tenant way.

This project has some reference from the [prometheus label injector](https://github.com/prometheus-community/prom-label-proxy)

The proxy enforces the `namespace` label in a given PromQL query while providing a basic auth layer.

## What is it?

It is a simple golang proxy. It does basic auth, logs the requests, and serves as a Prometheus reverse proxy.

Actually, [Prometheus](https://github.com/prometheus/prometheus) does not check the auth of any request.
By itself, it does not provide any multi-tenant mechanism. So, if you have untrusted tenants,
you have to ensure a tenant uses its labels and does not use any other tenants' value.

### Requirements

To use this project, place the proxy in front of your [Prometheus server](https://github.com/prometheus/prometheus)
instance, configure the auth proxy configuration and run it.

### Run it

```bash
$ prometheus-multi-tenant-proxy run --prometheus-endpoint http://localhost:9090 --port 9091 --auth-config ./my-auth-config.yaml --reload-interval=5
```

Where:

- `--port`: Port used to expose this proxy.
- `--prometheus-endpoint`: URL of your Prometheus instance.
- `--auth-config`: Authentication configuration file path.
- `--reload-interval`: Interval in minutes to reload the auth config file.

#### Configure the proxy

The auth configuration is straightforward. Just create a YAML file `my-auth-config.yaml` with the following structure:

```golang
// Authn Contains a list of users
type Authn struct {
	Users []User `yaml:"users"`
}

// User Identifies a user including the tenant
type User struct {
	Username  string `yaml:"username"`
	Password  string `yaml:"password"`
	Namespace string `yaml:"namespace"`
}
```

An example is available at [configs/multiple.user.yaml](configs/multiple.user.yaml) file:

```yaml
users:
  - username: User-a
    password: pass-a
    namespace: tenant-a
  - username: User-b
    password: pass-b
    namespace: tenant-b
```

A tenant can contain multiple users. But a user is tied to a simple tenant.

## Build it

If you want to build it from this repository, follow the instructions bellow:

```bash
$ docker run -it --entrypoint /bin/bash --rm golang:1.17-buster
root@6985c5523ed0:/go# git clone https://github.com/k8spin/prometheus-multi-tenant-proxy.git
Cloning into 'prometheus-multi-tenant-proxy'...
remote: Enumerating objects: 297, done.
remote: Counting objects: 100% (85/85), done.
remote: Compressing objects: 100% (42/42), done.
remote: Total 297 (delta 42), reused 57 (delta 37), pack-reused 212
Receiving objects: 100% (297/297), 209.10 KiB | 376.00 KiB/s, done.
Resolving deltas: 100% (120/120), done.
root@6985c5523ed0:/go# cd prometheus-multi-tenant-proxy/cmd/prometheus-multi-tenant-proxy/
root@6985c5523ed0:/go# go build
go: downloading github.com/urfave/cli/v2 v2.11.1
go: downloading github.com/prometheus/prometheus v0.35.0
go: downloading github.com/prometheus-community/prom-label-proxy v0.5.0
go: downloading gopkg.in/yaml.v2 v2.4.0
go: downloading github.com/pkg/errors v0.9.1
go: downloading github.com/efficientgo/tools/core v0.0.0-20220225185207-fe763185946b
go: downloading github.com/go-openapi/runtime v0.24.1
go: downloading github.com/go-openapi/strfmt v0.21.3
go: downloading github.com/prometheus/alertmanager v0.24.0
go: downloading github.com/asaskevich/govalidator v0.0.0-20210307081110-f21760c49a8d
go: downloading github.com/go-openapi/errors v0.20.2
go: downloading github.com/mitchellh/mapstructure v1.5.0
go: downloading github.com/oklog/ulid v1.3.1
go: downloading go.mongodb.org/mongo-driver v1.10.0
go: downloading github.com/opentracing/opentracing-go v1.2.0
go: downloading github.com/cpuguy83/go-md2man/v2 v2.0.2
go: downloading github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673
go: downloading github.com/prometheus/common v0.37.0
go: downloading github.com/go-openapi/swag v0.21.1
go: downloading github.com/go-openapi/validate v0.22.0
go: downloading github.com/go-openapi/analysis v0.21.4
go: downloading github.com/go-openapi/loads v0.21.1
go: downloading github.com/go-openapi/spec v0.20.6
go: downloading github.com/cespare/xxhash/v2 v2.1.2
go: downloading github.com/grafana/regexp v0.0.0-20220304095617-2e8d9baf4ac2
go: downloading github.com/russross/blackfriday/v2 v2.1.0
go: downloading github.com/mailru/easyjson v0.7.7
go: downloading github.com/go-openapi/jsonpointer v0.19.5
go: downloading github.com/go-kit/log v0.2.1
go: downloading github.com/josharian/intern v1.0.0
go: downloading github.com/go-openapi/jsonreference v0.20.0
go: downloading github.com/dennwc/varint v1.0.0
go: downloading github.com/prometheus/client_golang v1.12.2
go: downloading go.uber.org/atomic v1.9.0
go: downloading github.com/stretchr/testify v1.8.0
go: downloading github.com/go-logfmt/logfmt v0.5.1
go: downloading golang.org/x/sys v0.0.0-20220722155257-8c9f86f7a55f
go: downloading go.uber.org/goleak v1.1.12
go: downloading github.com/davecgh/go-spew v1.1.1
go: downloading github.com/pmezard/go-difflib v1.0.0
go: downloading gopkg.in/yaml.v3 v3.0.1
go: downloading github.com/prometheus/client_model v0.2.0
go: downloading github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369
go: downloading github.com/golang/protobuf v1.5.2
go: downloading github.com/beorn7/perks v1.0.1
go: downloading github.com/prometheus/procfs v0.7.3
go: downloading google.golang.org/protobuf v1.28.0
root@6985c5523ed0:/go# ./prometheus-multi-tenant-proxy
NAME:
   Prometheus multi-tenant proxy - Makes your Prometheus server multi tenant

USAGE:
   prometheus-multi-tenant-proxy [global options] command [command options] [arguments...]

VERSION:
   dev

AUTHORS:
   Angel Barrera <angel@k8spin.cloud>
   Pau Rosello <pau@k8spin.cloud>

COMMANDS:
   run      Runs the Prometheus multi-tenant proxy
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help (default: false)
   --version, -v  print the version (default: false)
```

### Build the container image

If you want to build a container image with this proxy, run:

```bash
$ docker build -t prometheus-multi-tenant-proxy:local -f build/package/Dockerfile .
```

After built, just run it:

```bash
$ docker run --rm prometheus-multi-tenant-proxy:local
```

## Using this project at work or in production?

See [ADOPTERS.md](ADOPTERS.md) for what companies are doing with this project today.

## License

The scripts and documentation in this project are released under the [GNU GPLv3](LICENSE)
