
Prometheus-multi-tenant-proxy
===========

K8Spin - Prometheus multi-tenant proxy


## Configuration

The following table lists the configurable parameters of the Prometheus-multi-tenant-proxy chart and their default values.

| Parameter                | Description             | Default        |
| ------------------------ | ----------------------- | -------------- |
| `replicaCount` |  | `1` |
| `image.repository` |  | `"ghcr.io/k8spin/prometheus-multi-tenant-proxy"` |
| `image.pullPolicy` |  | `"IfNotPresent"` |
| `image.tag` |  | `""` |
| `imagePullSecrets` |  | `[]` |
| `nameOverride` |  | `""` |
| `fullnameOverride` |  | `""` |
| `serviceAccount.create` |  | `true` |
| `serviceAccount.annotations` |  | `{}` |
| `serviceAccount.name` |  | `""` |
| `podAnnotations` |  | `{}` |
| `podSecurityContext.fsGroup` |  | `2000` |
| `securityContext.capabilities.drop` |  | `["ALL"]` |
| `securityContext.readOnlyRootFilesystem` |  | `true` |
| `securityContext.runAsNonRoot` |  | `true` |
| `securityContext.runAsUser` |  | `1000` |
| `service.type` |  | `"ClusterIP"` |
| `service.port` |  | `80` |
| `ingress.enabled` |  | `false` |
| `ingress.className` |  | `""` |
| `ingress.annotations` |  | `{}` |
| `ingress.hosts` |  | `[{"host": "chart-example.local", "paths": [{"path": "/", "pathType": "ImplementationSpecific"}]}]` |
| `ingress.tls` |  | `[]` |
| `resources` |  | `{}` |
| `autoscaling.enabled` |  | `false` |
| `autoscaling.minReplicas` |  | `1` |
| `autoscaling.maxReplicas` |  | `100` |
| `autoscaling.targetCPUUtilizationPercentage` |  | `80` |
| `nodeSelector` |  | `{}` |
| `tolerations` |  | `[]` |
| `affinity` |  | `{}` |
| `proxy.port` |  | `9092` |
| `proxy.prometheusEndpoint` |  | `""` |
| `proxy.extraArgs` |  | `[]` |
| `proxy.extraEnv` |  | `[]` |
| `proxy.auth.type` | basic or jwt | `"basic"` |
| `proxy.auth.jwt.url` | URL/Path to the JWT configuration | `""` |
| `proxy.auth.basic.createSecret` |  | `true` |
| `proxy.auth.basic.secretName` | In use only if createSecret is false | `"prometheus-multi-tenant-proxy"` |





