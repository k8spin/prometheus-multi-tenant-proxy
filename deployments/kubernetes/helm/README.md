# CNCT Version to build and deploy

## Chart version:

    Note: the original chart has 0.0.0 in the version field.
    
    You must edit the version numbers in `Chart.yaml`
    
    2.80.0 <- using 80 for cnct

## Sequence of commands
```

docker pull --platform linux/amd64 ghcr.io/k8spin/prometheus-multi-tenant-proxy:latest
docker image inspect   --format "{{.ID}} {{.RepoTags}} {{.Architecture}}" ghcr.io/k8spin/prometheus-multi-tenant-proxy:latest
docker images
docker tag ghcr.io/k8spin/prometheus-multi-tenant-proxy:latest obv-miefxkks.scr.kr-west.scp-in.com/images/k8spin/prometheus-multi-tenant-proxy:2.80.0
docker images
docker push obv-miefxkks.scr.kr-west.scp-in.com/images/k8spin/prometheus-multi-tenant-proxy:2.80.0
helm version
vi prometheus-multi-tenant-proxy/values.yaml
helm --debug package prometheus-multi-tenant-proxy/
helm --debug push  ./prometheus-multi-tenant-proxy-2.80.0.tgz oci://obv-miefxkks.scr.kr-west.scp-in.com/charts/k8spin
cd download
helm --debug pull oci://obv-miefxkks.scr.kr-west.scp-in.com/charts/k8spin/prometheus-multi-tenant-proxy:2.80.0
docker pull obv-miefxkks.scr.kr-west.scp-in.com/images/k8spin/prometheus-multi-tenant-proxy:2.80.0
docker image inspect   --format "{{.ID}} {{.RepoTags}} {{.Architecture}}" obv-miefxkks.scr.kr-west.scp-in.com/images/k8spin/prometheus-multi-tenant-proxy:2.80.0
```