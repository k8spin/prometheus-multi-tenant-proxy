# Prometheus — Multi-tenant proxy

![Build Status](https://github.com/k8spin/prometheus-multi-tenant-proxy/actions/workflows/release.yml/badge.svg)
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)

[![Twitter](https://img.shields.io/twitter/url/https/twitter.com/k8spin.svg?style=social&label=Follow%20%40k8spin)](https://twitter.com/k8spin)
[![Join the chat at https://slack.kubernetes.io](https://img.shields.io/badge/style-register-green.svg?style=social&label=Slack)](https://slack.kubernetes.io)

------

This project aims to make it easy to deploy a [Prometheus Server](https://github.com/prometheus/prometheus)
in a multi-tenant way.

This project has some reference from the [prometheus label injector](https://github.com/prometheus-community/prom-label-proxy)

The proxy enforces the `namespace` label in a given PromQL query while providing a basic auth layer.

## What is it?

It is a simple golang proxy. It does basic or JWT auth, logs the requests, and serves as a Prometheus reverse proxy.

Actually, [Prometheus](https://github.com/prometheus/prometheus) does not check the auth of any request.
By itself, it does not provide any multi-tenant mechanism. So, if you have untrusted tenants,
you have to ensure a tenant uses its labels and does not use any other tenants' value.

For more security, only specific endpoints are proxied by default: `/api/v1/series`, `/api/v1/query`,
and `/api/v1/query_range` (see `--protected-endpoints`).

The proxy also supports Amazon Managed Service for Prometheus.

### Requirements

To use this project, place the proxy in front of your [Prometheus server](https://github.com/prometheus/prometheus)
instance, configure the auth proxy configuration and run it.

### Run it

```bash
$ prometheus-multi-tenant-proxy run \
  --prometheus-endpoint http://localhost:9090 \
  --port 9091 \
  --auth-config ./my-auth-config.yaml \
  --reload-interval=5 \
  --unprotected-endpoints /-/healthy,/-/ready
```

Available arguments // environment variables to the `run` command:

- `--port` // `PROM_PROXY_PORT`: Port used to expose this proxy.
- `--prometheus-endpoint` // `PROM_PROXY_PROMETHEUS_ENDPOINT`: URL of your Prometheus instance.
- `--reload-interval` // `PROM_PROXY_RELOAD_INTERVAL`: Interval in minutes to reload the auth config file.
- `--unprotected-endpoints` // `PROM_PROXY_UNPROTECTED_ENDPOINTS`: Comma separated list of endpoints that do not require authentication.
- `--protected-endpoints` // `PROM_PROXY_PROTECTED_ENDPOINTS`: Comma separated list of endpoints that are allowed after authentication.
   Pass an empty string to turn it off (i.e. to allow all endpoints).
- `--auth-type` // `PROM_PROXY_AUTH_TYPE`: Type of authentication to use, one of `basic`,  `jwt`
- `--auth-config` // `PROM_PROXY_AUTH_CONFIG`: Authentication configuration.
   * for `basic` authentication: path to a configuration file following the *Authn structure*
   * for `jwt` authentication: either a path or an URL to a json containing a *Json Web Keys Set (JWKS)*
- `--aws` // `PROM_PROXY_USE_AWS`: See below.

Use `prometheus-multi-tenant-proxy run --help` for more information.

#### Configure the proxy for basic authentication

The auth configuration is straightforward. Just create a YAML file `my-auth-config.yaml` with the following structure:

```golang
// Authn Contains a list of users
type Authn struct {
	Users []User `yaml:"users"`
}

// User Identifies a user including the tenant
type User struct {
	Username   string              `yaml:"username"`
	Password   string              `yaml:"password"`
	Namespace  string              `yaml:"namespace"`
	Namespaces []string            `yaml:"namespaces"`
	Labels     map[string][]string `yaml:"labels"`
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

Tenant definition usually contains a set of labels. Starting from v1.7.0 it's possible to add these labels to a new `labels`
section to the user definition to inject these labels on queries for that user.

Example available at [configs/sample.labels.yaml](configs/sample.labels.yaml) file:

```yaml
users:
  - username: Happy
    password: Prometheus
    labels:
      app:
        - happy
        - sad
      team:
        - america
  - username: Sad
    password: Prometheus
    labels:
      namespace:
        - kube-system
        - monitoring
  - username: bored
    password: Prometheus
    namespaces:
      - default
      - kube-system
    labels:
      dep:
        - system
```

#### Configure the proxy for JWT authentication

Under the hood, the proxy uses [keyfunc](https://github.com/MicahParks/keyfunc) to load
keys (in JWKS format), and [go-jwt](https://github.com/golang-jwt/jwt) for validating JWT tokens.

The **Json Web Keys Set (JWKS)** can be loaded either from a file or an URL,
and will be reloaded automatically following the `--reload-interval` parameter.

An example of a valid JWKS containing both an HS256 (hmac, symmetric) and an RS256 (rsa, asymmetric) key is available
at [internal/app/prometheus-multi-tenant-proxy/.jwks_example.json](internal/app/prometheus-multi-tenant-proxy/.jwks_example.json).
More examples are provided in the [keyfunc](https://github.com/MicahParks/keyfunc) readme.
You can also use [mkjwk.org](https://mkjwk.org) to generate valid JWKs.

Once the proxy is aware of one or more JWKS keys, it is ready to authorize requests based on signed JWT tokens.
The **token** is extracted from one of two locations with the given precedence:

1. the `Authorization` header, in the form `Authorization: Bearer <TOKEN>`, or, if not present,
2. the `Token` header, in the form `Token: <TOKEN>`.

For the token to be valid, it must:

* contain a `kid` (key ID) in the header that matches the kid of a known key in the JWKS,
* contain a claim in the payload called `namespaces`, with zero or more values. For example:
  ```json
  {
    "namespaces": ["foo", "bar"]
  }
  ```
* contain a claim in the payload called `labels`, with zero or more values. For example:
  ```json
  {
    "labels": {
      "app": ["happy", "sad"],
      "team": ["america"]
    }
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

#### Proxy to Amazon Managed Service for Prometheus

All requests to an AWS managed prometheus service need a signature in the `Authorization` header,
which is calculated based on the request URL, headers, and body.
Since the proxy modifies the request on the fly, any existing signature will
be invalidated.
This is why prometheus-multi-tenant-proxy incorporates AWS signature v4.

To enable AWS signature, use either the `--aws` flag or set the environment variable
`PROM_PROXY_USE_AWS=true`.

The credentials used for signing are taken from environment variables, such as `AWS_ACCESS_KEY` and `AWS_SECRET_KEY`
(see [AWS credentials environment variables](
https://docs.aws.amazon.com/sdk-for-php/v3/developer-guide/guide_credentials_environment.html) for a full list).
In case your prometheus service doesn't live in the `us-east-1` region, you will also have to set
either `AWS_REGION` or `AWS_DEFAULT_REGION` to the region's shorthand.
Note that the AWS service is always set to `aps`, which is the shorthand for AWS Prometheus Service.

For more information, see:

  * [What is Amazon Managed Service for Prometheus?](
    https://docs.aws.amazon.com/prometheus/latest/userguide/what-is-Amazon-Managed-Service-Prometheus.html)
  * [Signature Calculations for the Authorization Header: Transferring Payload in a Single Chunk (AWS Signature Version 4)](
    https://docs.aws.amazon.com/AmazonS3/latest/API/sig-v4-header-based-auth.html)
  * [Using credentials from environment variables](https://docs.aws.amazon.com/sdk-for-php/v3/developer-guide/guide_credentials_environment.html)


#### Namespaces or labels

The proxy can be configured to use either namespaces and/or labels to query Prometheus.
At least one must be configured, otherwise the proxy will not proxy the query to Prometheus.
*(It could lead to a security issue if the proxy is not configured to use namespaces or labels)*

##### Breaking Change in [v2.0.0](https://github.com/k8spin/prometheus-multi-tenant-proxy/releases/tag/v2.0.0): Update `map[string]string` to `map[string][]string` for the labels map values

What Changed: Previously, the map only allowed a single string value per key:

```
// Old implementation (prior to version 2.0.0)
labels:
   app: happy
   team: america
```
Now, the map allows each key to store a slice of strings (multiple values):

```
// New implementation (starting from version 2.0.0)
labels:
   app:
     - happy
     - sad
   team:
     - america
```
### Deploy on Kubernetes using Helm

The proxy can be deployed on Kubernetes using Helm. The Helm chart is available at [k8spin/prometheus-multi-tenant-proxy](https://k8spin.github.io/prometheus-multi-tenant-proxy). Find the chart's documentation on its [README.md](deployments/kubernetes/helm/prometheus-multi-tenant-proxy/README.md).

TL;DR:

```bash
$ helm repo add k8spin-prometheus-multi-tenant-proxy https://k8spin.github.io/prometheus-multi-tenant-proxy
$ helm repo update
$ helm upgrade --install prometheus-multi-tenant-proxy k8spin-prometheus-multi-tenant-proxy/prometheus-multi-tenant-proxy --set proxy.prometheusEndpoint=http://prometheus.monitoring.svc.cluster.local:9090
```

#### Example using flux

```yaml
---
apiVersion: source.toolkit.fluxcd.io/v1beta2
kind: HelmRepository
metadata:
  name: prometheus-multi-tenant-proxy
  namespace: flux-system
  labels:
    phase: seed
spec:
  interval: 1m0s
  url: https://k8spin.github.io/prometheus-multi-tenant-proxy
---
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  name: prometheus-multi-tenant-proxy
  namespace: flux-system
spec:
  timeout: 30m
  install:
    remediation:
      retries: 3
  upgrade:
    remediation:
      retries: 3
  interval: 1m
  chart:
    spec:
      chart: prometheus-multi-tenant-proxy
      version: "1.10.0"
      sourceRef:
        kind: HelmRepository
        name: prometheus-multi-tenant-proxy
        namespace: flux-system
      interval: 1m
  releaseName: prometheus-multi-tenant-proxy
  targetNamespace: monitoring
  storageNamespace: monitoring
  valuesFrom: []
  values:
    proxy:
      prometheusEndpoint: http://prometheus.monitoring.svc.cluster.local:9090
      auth:
        basic:
          authn: |
            users:
              - username: User-a
                password: pass-a
                namespace: tenant-a
              - username: User-b
                password: pass-b
                namespace: tenant-b
```

## Build it

If you want to build it from this repository, follow the instructions below:

```bash
$ docker run -it --entrypoint /bin/bash --rm golang:1.23.1-bookworm
root@9b2da74fb4b8:/go# git clone https://github.com/k8spin/prometheus-multi-tenant-proxy.git
Cloning into 'prometheus-multi-tenant-proxy'...
remote: Enumerating objects: 877, done.
remote: Counting objects: 100% (235/235), done.
remote: Compressing objects: 100% (125/125), done.
remote: Total 877 (delta 147), reused 144 (delta 104), pack-reused 642 (from 1)
Receiving objects: 100% (877/877), 637.41 KiB | 4.25 MiB/s, done.
Resolving deltas: 100% (466/466), done.
root@9b2da74fb4b8:/go# cd prometheus-multi-tenant-proxy/cmd/prometheus-multi-tenant-proxy/
root@9b2da74fb4b8:/go/prometheus-multi-tenant-proxy/cmd/prometheus-multi-tenant-proxy# go build
go: downloading github.com/urfave/cli/v2 v2.27.4
go: downloading github.com/MicahParks/keyfunc/v2 v2.1.0
go: downloading github.com/aws/aws-sdk-go v1.55.5
go: downloading github.com/golang-jwt/jwt/v5 v5.2.1
go: downloading github.com/prometheus-community/prom-label-proxy v0.11.0
go: downloading github.com/prometheus/prometheus v0.54.1
go: downloading gopkg.in/yaml.v3 v3.0.1
go: downloading github.com/efficientgo/core v1.0.0-rc.2
go: downloading github.com/go-openapi/runtime v0.28.0
go: downloading github.com/go-openapi/strfmt v0.23.0
go: downloading github.com/metalmatze/signal v0.0.0-20210307161603-1c9aa721a97a
go: downloading github.com/prometheus/alertmanager v0.27.0
go: downloading github.com/prometheus/client_golang v1.19.1
go: downloading github.com/cpuguy83/go-md2man/v2 v2.0.4
go: downloading github.com/xrash/smetrics v0.0.0-20240521201337-686a1a2994c1
go: downloading github.com/cespare/xxhash/v2 v2.3.0
go: downloading github.com/grafana/regexp v0.0.0-20240518133315-a468a5bfb3bc
go: downloading github.com/prometheus/common v0.55.0
go: downloading golang.org/x/text v0.16.0
go: downloading github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2
go: downloading github.com/go-openapi/errors v0.22.0
go: downloading github.com/google/uuid v1.6.0
go: downloading github.com/mitchellh/mapstructure v1.5.0
go: downloading github.com/oklog/ulid v1.3.1
go: downloading go.mongodb.org/mongo-driver v1.14.0
go: downloading github.com/opentracing/opentracing-go v1.2.0
go: downloading go.opentelemetry.io/otel v1.28.0
go: downloading go.opentelemetry.io/otel/trace v1.28.0
go: downloading github.com/go-openapi/swag v0.23.0
go: downloading github.com/go-openapi/validate v0.24.0
go: downloading github.com/russross/blackfriday/v2 v2.1.0
go: downloading github.com/beorn7/perks v1.0.1
go: downloading github.com/prometheus/client_model v0.6.1
go: downloading github.com/prometheus/procfs v0.15.1
go: downloading google.golang.org/protobuf v1.34.2
go: downloading github.com/go-kit/log v0.2.1
go: downloading golang.org/x/sync v0.7.0
go: downloading github.com/go-openapi/analysis v0.23.0
go: downloading github.com/go-openapi/loads v0.22.0
go: downloading github.com/go-openapi/spec v0.21.0
go: downloading github.com/go-logr/logr v1.4.2
go: downloading go.opentelemetry.io/otel/metric v1.28.0
go: downloading github.com/mailru/easyjson v0.7.7
go: downloading github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822
go: downloading github.com/go-openapi/jsonpointer v0.21.0
go: downloading golang.org/x/sys v0.22.0
go: downloading github.com/go-logfmt/logfmt v0.6.0
go: downloading github.com/dennwc/varint v1.0.0
go: downloading go.uber.org/atomic v1.11.0
go: downloading github.com/go-logr/stdr v1.2.2
go: downloading github.com/go-openapi/jsonreference v0.21.0
go: downloading github.com/josharian/intern v1.0.0
go: downloading github.com/jmespath/go-jmespath v0.4.0
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
