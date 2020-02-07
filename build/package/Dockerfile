FROM golang:1.12-alpine as builder

ENV GO111MODULE=on
ENV CGO_ENABLED=0

RUN apk add  -U --no-cache git ca-certificates && \
    mkdir -p src/github.com/angelbarrera92/prometheus-multi-tenant-proxy

WORKDIR /go/src/github.com/angelbarrera92/prometheus-multi-tenant-proxy

COPY go.mod go.mod
COPY go.sum go.sum
COPY cmd cmd
COPY internal internal

RUN cd cmd/prometheus-multi-tenant-proxy && \
    go build .

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/github.com/angelbarrera92/prometheus-multi-tenant-proxy/cmd/prometheus-multi-tenant-proxy/prometheus-multi-tenant-proxy /prometheus-multi-tenant-proxy

ENTRYPOINT [ "/prometheus-multi-tenant-proxy" ]