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
$ prometheus-multi-tenant-proxy run --prometheus-endpoint http://localhost:9090 --port 9091 --auth-config ./my-auth-config.yaml --reload-interval=5 --unprotected-endpoints /-/healthy,/-/ready
```

Where:

- `--port`: Port used to expose this proxy.
- `--prometheus-endpoint`: URL of your Prometheus instance.
- `--reload-interval`: Interval in minutes to reload the auth config file.
- `--unprotected-endpoints`: Comma separated list of endpoints that do not require authentication.
- `--auth-type`: Type of authentication to use, one of `basic`,  `jwt`
- `--auth-config`: Authentication configuration.
   * for `basic` authentication: path to a configuration file following the *Authn structure*
   * for `jwt` authentication: either a path or an URL to a json containing a *Json Web Keys Set (JWKS)*

#### Configure the proxy for basic authentication

The auth configuration is straightforward. Just create a YAML file `my-auth-config.yaml` with the following structure:

```golang
// Authn Contains a list of users
type Authn struct {
	Users []User `yaml:"users"`
}

// User Identifies a user including the tenant
type User struct {
	Username   string   `yaml:"username"`
	Password   string   `yaml:"password"`
	Namespace  string   `yaml:"namespace"`
	Namespaces []string `yaml:"namespaces"`
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

or if you need to allow multiple namespaces for a single user,
an example is available at [configs/multiple.namespaces.yaml](configs/multiple.namespaces.yaml) file:

```yaml
users:
  - username: Happy
    password: Prometheus
    namespace: default
  - username: Sad
    password: Prometheus
    namespace: kube-system
  - username: Multiple
    password: Namespaces
    namespace: monitoring
    namespaces:
      - default
      - kube-system
      - kube-public
  - username: Multiple
    password: NamespacesWithoutNamespace
    namespaces:
      - default
      - kube-system
      - kube-public
```

A tenant can contain multiple users. But a user is tied to a single tenant.

#### Configure the proxy for JWT authentication

Under the hood, the proxy uses [keyfunc](https://github.com/MicahParks/keyfunc) to load
keys (in JWKS format), and [go-jwt](https://github.com/golang-jwt/jwt) for validating JWT tokens.

The **Json Web Keys Set (JWKS)** can be loaded either from a file or an URL,
and will be reloaded automatically following the `--reload-interval` parameter.

An example of a valid JWKS containing both an HS256 (hmac, symmetric) and an RS256 (rsa, asymmetric) key is available
at [internal/app/prometheus-multi-tenant-proxy/.jwks_example.json](#internal/app/prometheus-multi-tenant-proxy/.jwks_example.json).
More examples are provided in the [keyfunc](https://github.com/MicahParks/keyfunc) readme.
You can also use [mkjwk.org](https://mkjwk.org) to generate valid JWKs.

Once the proxy is aware of one or more JWKS keys, it is ready to authorize requests based on signed JWT tokens.
The **token** is extracted from one of two locations with the given precedence:

1. the `Authorization` header, in the form `Authorization: Bearer <TOKEN>`, or, if not present,
2. the `Token` header, in the form `Token: <TOKEN>`.

For the token to be valid, it must:

* contain a `kid` (key ID) in the header that matches the kid of a known key in the JWKS,
* contain a claim in the payload called `namespaces`, with one or more values. For example:
  ```json
  {
    "namespaces": ["foo", "bar"]
  }
  ```
* have been signed with the key in the JWKS matching the `kid` found in the JWT header.

To test the proxy using JWT tokens, you can use the `.jwks_example.json` file above to run
the proxy and generate a JWT token using [jwt.io](https://jwt.io). Ensure you chose the `HS256` algorithm and
paste the following token:
```json
eyJhbGciOiJIUzI1NiIsImtpZCI6ImhtYWMta2V5In0.eyJuYW1lc3BhY2VzIjpbInByb21ldGhldXMiLCJhcHAtMSJdfQ.aMVibvV_meujcnRA1pgnSjBojtzvteZSf2xvq2MwZgc
```
To verify the signature, replace `your-256-bit-secret` with `lala` in the "verify signature" section.
You can now use curl, for example:
```bash
curl -H "Authorization: Bearer $TOKEN" http://localhost:9092/api/v1/query\?query\=net_conntrack_dialer_conn_attempted_total
```

## Build it

If you want to build it from this repository, follow the instructions below:

```bash
$ docker run -it --entrypoint /bin/bash --rm golang:1.18.8-buster
root@9b2da74fb4b8:/go# git clone https://github.com/k8spin/prometheus-multi-tenant-proxy.git
Cloning into 'prometheus-multi-tenant-proxy'...
remote: Enumerating objects: 310, done.
remote: Counting objects: 100% (15/15), done.
remote: Compressing objects: 100% (9/9), done.
remote: Total 310 (delta 6), reused 14 (delta 6), pack-reused 295
Receiving objects: 100% (310/310), 255.71 KiB | 1.66 MiB/s, done.
Resolving deltas: 100% (122/122), done.
root@9b2da74fb4b8:/go# cd prometheus-multi-tenant-proxy/cmd/prometheus-multi-tenant-proxy/
root@9b2da74fb4b8:/go/prometheus-multi-tenant-proxy/cmd/prometheus-multi-tenant-proxy# go build
go: downloading github.com/urfave/cli/v2 v2.24.3
go: downloading github.com/prometheus/prometheus v0.42.0
go: downloading github.com/prometheus-community/prom-label-proxy v0.6.0
go: downloading gopkg.in/yaml.v2 v2.4.0
go: downloading github.com/efficientgo/tools/core v0.0.0-20220225185207-fe763185946b
go: downloading github.com/go-openapi/runtime v0.25.0
go: downloading github.com/go-openapi/strfmt v0.21.3
go: downloading github.com/metalmatze/signal v0.0.0-20210307161603-1c9aa721a97a
go: downloading github.com/pkg/errors v0.9.1
go: downloading github.com/prometheus/alertmanager v0.25.0
go: downloading github.com/prometheus/client_golang v1.14.0
go: downloading github.com/cpuguy83/go-md2man/v2 v2.0.2
go: downloading github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673
go: downloading github.com/opentracing/opentracing-go v1.2.0
go: downloading go.opentelemetry.io/otel v1.11.2
go: downloading go.opentelemetry.io/otel/trace v1.11.2
go: downloading github.com/asaskevich/govalidator v0.0.0-20210307081110-f21760c49a8d
go: downloading github.com/go-openapi/errors v0.20.3
go: downloading github.com/mitchellh/mapstructure v1.5.0
go: downloading github.com/oklog/ulid v1.3.1
go: downloading go.mongodb.org/mongo-driver v1.11.0
go: downloading github.com/cespare/xxhash/v2 v2.2.0
go: downloading github.com/grafana/regexp v0.0.0-20221122212121-6b5c0a4cb7fd
go: downloading github.com/prometheus/common v0.39.0
go: downloading github.com/go-openapi/swag v0.22.3
go: downloading github.com/go-openapi/validate v0.22.0
go: downloading github.com/go-openapi/analysis v0.21.4
go: downloading github.com/go-openapi/loads v0.21.2
go: downloading github.com/go-openapi/spec v0.20.7
go: downloading github.com/russross/blackfriday/v2 v2.1.0
go: downloading github.com/beorn7/perks v1.0.1
go: downloading github.com/golang/protobuf v1.5.2
go: downloading github.com/prometheus/client_model v0.3.0
go: downloading github.com/prometheus/procfs v0.8.0
go: downloading google.golang.org/protobuf v1.28.1
go: downloading github.com/go-logr/logr v1.2.3
go: downloading github.com/go-kit/log v0.2.1
go: downloading golang.org/x/exp v0.0.0-20230124195608-d38c7dcee874
go: downloading github.com/mailru/easyjson v0.7.7
go: downloading gopkg.in/yaml.v3 v3.0.1
go: downloading github.com/go-openapi/jsonpointer v0.19.5
go: downloading github.com/go-openapi/jsonreference v0.20.0
go: downloading github.com/matttproud/golang_protobuf_extensions v1.0.4
go: downloading github.com/go-logr/stdr v1.2.2
go: downloading golang.org/x/sys v0.4.0
go: downloading github.com/dennwc/varint v1.0.0
go: downloading go.uber.org/atomic v1.10.0
go: downloading github.com/stretchr/testify v1.8.1
go: downloading github.com/go-logfmt/logfmt v0.5.1
go: downloading github.com/josharian/intern v1.0.0
go: downloading github.com/gogo/protobuf v1.3.2
go: downloading github.com/davecgh/go-spew v1.1.1
go: downloading go.uber.org/goleak v1.2.0
go: downloading github.com/pmezard/go-difflib v1.0.0
root@9b2da74fb4b8:/go/prometheus-multi-tenant-proxy/cmd/prometheus-multi-tenant-proxy# ./prometheus-multi-tenant-proxy
NAME:
   Prometheus multi-tenant proxy - Makes your Prometheus server multi tenant

USAGE:
   Prometheus multi-tenant proxy [global options] command [command options] [arguments...]

VERSION:
   dev

AUTHORS:
   Angel Barrera <angel@k8spin.cloud>
   Pau Rosello <pau@k8spin.cloud>

COMMANDS:
   run      Runs the Prometheus multi-tenant proxy
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version
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
