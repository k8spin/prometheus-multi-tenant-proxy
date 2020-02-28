# Prometheus - Multi tenant proxy

This project has been developed to make easy to deploy a [Prometheus Server](https://github.com/prometheus/prometheus)
in a multi-tenant way.

This works is based on the [prometheus label injector from RedHat](https://github.com/openshift/prom-label-proxy)

This proxy enforces the `namespace` label in a given PromQL query while providing a basic auth layer.

## What is it?

It is a basic golang proxy. It does basic auth, logs the requests and serves as a prometheus reverse proxy.

Actually, [Prometheus](https://github.com/prometheus/prometheus) does not check the auth of any request.
By it self, it does not provide any kind of multi-tenant mechanism. So, if you have untrusted tenants,
you have to ensure a tenant uses it's own labels and does not use any value from other tenants.

### Requirements

To use this proxy, put the proxy in front of your [Prometheus server](https://github.com/prometheus/prometheus)
instance, configure the auth proxy configuration, and run it.

### Run it

```bash
$ prometheus-multi-tenant-proxy run --prometheus-endpoint http://localhost:9090 --port 9091 --auth-config ./my-auth-config.yaml
```

Where:

- `--port`: Port used to expose this proxy.
- `--prometheus-endpoint`: URL of your prometheus instance.
- `--auth-config`: Authentication configuration file path.

#### Configure the proxy

The auth configuration is very simple. Just create a yaml file `my-auth-config.yaml` with the following structure:

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

A tenant can contains multiple users. But a user is tied to a simple tenant.

## Build it

If you want to build it from this repository, follow the instructions bellow:

```bash
$ docker run -it --entrypoint /bin/bash --rm golang:latest
root@6985c5523ed0:/go# git clone https://github.com/k8spin/prometheus-multi-tenant-proxy.git
Cloning into 'prometheus-multi-tenant-proxy'...
remote: Enumerating objects: 96, done.
remote: Counting objects: 100% (96/96), done.
remote: Compressing objects: 100% (54/54), done.
remote: Total 96 (delta 31), reused 87 (delta 22), pack-reused 0
Unpacking objects: 100% (96/96), done.
root@6985c5523ed0:/go# cd prometheus-multi-tenant-proxy/cmd/prometheus-multi-tenant-proxy/
root@6985c5523ed0:/go# go build
go: downloading github.com/urfave/cli v1.21.0
go: downloading github.com/prometheus/prometheus v1.8.2-0.20200106144642-d9613e5c466c
go: downloading gopkg.in/yaml.v2 v2.2.5
go: downloading github.com/prometheus/common v0.7.0
go: downloading github.com/go-kit/kit v0.9.0
go: downloading github.com/edsrzf/mmap-go v1.0.0
go: downloading github.com/opentracing/opentracing-go v1.1.0
go: downloading github.com/cespare/xxhash v1.1.0
go: downloading github.com/prometheus/client_golang v1.2.0
go: downloading github.com/pkg/errors v0.8.1
go: downloading github.com/golang/snappy v0.0.1
go: downloading github.com/oklog/ulid v1.3.1
go: downloading golang.org/x/sys v0.0.0-20191010194322-b09406accb47
go: downloading golang.org/x/sync v0.0.0-20190423024810-112230192c58
go: downloading github.com/alecthomas/units v0.0.0-20190717042225-c3de453c63f4
go: downloading github.com/prometheus/procfs v0.0.5
go: downloading github.com/cespare/xxhash/v2 v2.1.0
go: downloading github.com/golang/protobuf v1.3.2
go: downloading github.com/beorn7/perks v1.0.1
go: downloading github.com/prometheus/client_model v0.0.0-20190812154241-14fe0d1b01d4
go: downloading github.com/go-logfmt/logfmt v0.4.0
go: downloading github.com/matttproud/golang_protobuf_extensions v1.0.1
root@6985c5523ed0:/go# ./prometheus-multi-tenant-proxy
NAME:
   Prometheus Multitenant Proxy - Makes your Prometheus server multi tenant

USAGE:
   prometheus-multi-tenant-proxy [global options] command [command options] [arguments...]

VERSION:
   dev

AUTHORS:
   Angel Barrera <angel@k8spin.cloud>
   Pau Rosello <pau@k8spin.cloud>

COMMANDS:
   run      Runs the Prometheus multi tenant proxy
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version
```

### Build the container image

If you want to build a container image with this proxy, simply run:

```bash
$ docker build -t prometheus-multi-tenant-proxy:local -f build/package/Dockerfile .
```

After built, just run it:

```bash
$ docker run --rm prometheus-multi-tenant-proxy:local
```
