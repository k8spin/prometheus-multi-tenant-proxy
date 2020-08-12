FROM golang:1.14-stretch as builder

ARG VERSION=dev
ARG COMMIT=none

ENV CGO_ENABLED=0

RUN apt-get update && apt-get install -y git ca-certificates && \
    mkdir -p /go/src/github.com/k8spin/prometheus-multi-tenant-proxy

WORKDIR /go/src/github.com/k8spin/prometheus-multi-tenant-proxy

COPY go.mod go.mod
COPY go.sum go.sum
COPY cmd cmd
COPY pkg pkg
COPY internal internal

RUN cd cmd/prometheus-multi-tenant-proxy && \
    go build -ldflags="-X 'main.version=${VERSION}' -X 'main.commit=${COMMIT}'"

FROM debian:stretch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/github.com/k8spin/prometheus-multi-tenant-proxy/cmd/prometheus-multi-tenant-proxy/prometheus-multi-tenant-proxy /prometheus-multi-tenant-proxy

ENTRYPOINT [ "/prometheus-multi-tenant-proxy" ]