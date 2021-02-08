# Prometheus — Multi-tenant proxy

![Build Status](https://action-badges.now.sh/k8spin/k8spin-operator)
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)

[![Twitter](https://img.shields.io/twitter/url/https/twitter.com/k8spin.svg?style=social&label=Follow%20%40k8spin)](https://twitter.com/k8spin)
[![Join the chat at https://slack.kubernetes.io](https://img.shields.io/badge/style-register-green.svg?style=social&label=Slack)](https://slack.kubernetes.io)

------

This project aims to make it easy to deploy a [Prometheus Server](https://github.com/prometheus/prometheus)
in a multi-tenant way.

This project has some reference from the [prometheus label injector from RedHat](https://github.com/openshift/prom-label-proxy)

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
$ prometheus-multi-tenant-proxy run --prometheus-endpoint http://localhost:9090 --port 9091 --auth-config ./my-auth-config.yaml
```

Where:

- `--port`: Port used to expose this proxy.
- `--prometheus-endpoint`: URL of your Prometheus instance.
- `--auth-config`: Authentication configuration file path.

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
$ docker run -it --entrypoint /bin/bash --rm golang:1.15.8-buster
root@6985c5523ed0:/go# git clone https://github.com/k8spin/prometheus-multi-tenant-proxy.git
Cloning into 'prometheus-multi-tenant-proxy'...
remote: Enumerating objects: 96, done.
remote: Counting objects: 100% (96/96), done.
remote: Compressing objects: 100% (54/54), done.
remote: Total 96 (delta 31), reused 87 (delta 22), pack-reused 0
Unpacking objects: 100% (96/96), done.
root@6985c5523ed0:/go# cd prometheus-multi-tenant-proxy/cmd/prometheus-multi-tenant-proxy/
root@6985c5523ed0:/go# go build
go: downloading github.com/urfave/cli v1.22.1
go: downloading github.com/prometheus/prometheus v1.8.2-0.20200507164740-ecee9c8abfd1
go: downloading github.com/prometheus-community/prom-label-proxy v0.2.1-0.20210129135803-4c30ca94e827
go: downloading gopkg.in/yaml.v2 v2.4.0
go: downloading github.com/urfave/cli/v2 v2.3.0
go: downloading github.com/prometheus/alertmanager v0.20.0
go: downloading github.com/go-openapi/runtime v0.19.15
go: downloading github.com/pkg/errors v0.9.1
go: downloading github.com/go-openapi/strfmt v0.19.5
go: downloading github.com/go-openapi/analysis v0.19.10
go: downloading github.com/go-openapi/loads v0.19.5
go: downloading github.com/mitchellh/mapstructure v1.2.2
go: downloading github.com/go-openapi/validate v0.19.8
go: downloading go.mongodb.org/mongo-driver v1.3.2
go: downloading github.com/go-openapi/swag v0.19.9
go: downloading github.com/go-openapi/spec v0.19.7
go: downloading github.com/prometheus/common v0.9.1
go: downloading github.com/cespare/xxhash v1.1.0
go: downloading github.com/go-openapi/errors v0.19.4
go: downloading github.com/cpuguy83/go-md2man/v2 v2.0.0-20190314233015-f79a8a8ca69d
go: downloading github.com/go-kit/kit v0.10.0
go: downloading golang.org/x/sys v0.0.0-20200420163511-1957bb5e6d1f
go: downloading github.com/go-openapi/jsonpointer v0.19.3
go: downloading github.com/asaskevich/govalidator v0.0.0-20200108200545-475eaeb16496
go: downloading github.com/mailru/easyjson v0.7.1
go: downloading github.com/russross/blackfriday/v2 v2.0.1
go: downloading github.com/go-openapi/jsonreference v0.19.3
go: downloading github.com/go-logfmt/logfmt v0.5.0
go: downloading github.com/shurcooL/sanitized_anchor_name v1.0.0
go: downloading github.com/PuerkitoBio/purell v1.1.1
go: downloading github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578
go: downloading golang.org/x/text v0.3.2
go: downloading golang.org/x/net v0.0.0-20200421231249-e086a090c8fd
go: downloading github.com/go-stack/stack v1.8.0
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
