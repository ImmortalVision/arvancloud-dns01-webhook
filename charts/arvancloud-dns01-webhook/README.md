# arvancloud-dns01-webhook Helm Chart

Helm chart for deploying the cert-manager DNS01 webhook for ArvanCloud.

## Prerequisites

- Kubernetes cluster
- cert-manager installed

## Install

```bash
helm upgrade --install arvancloud-dns01-webhook ./charts/arvancloud-dns01-webhook
```

## Example override

```bash
helm upgrade --install arvancloud-dns01-webhook ./charts/arvancloud-dns01-webhook \
  --set image.tag=v0.1.0 \
  --set groupName=acme.arvancloud.ir
```

## Uninstall

```bash
helm uninstall arvancloud-dns01-webhook -n arvancloud-dns01-webhook
```

## Values

| Value | Default | Description |
| --- | --- | --- |
| `groupName` | `acme.arvancloud.ir` | Webhook API group and `GROUP_NAME` env for cert-manager solver matching. |
| `image.repository` | `ghcr.io/immortalvision/arvancloud-dns01-webhook` | Webhook image repository. |
| `image.tag` | `latest` | Webhook image tag. |
| `namespace.create` | `true` | Create target namespace from chart. |
| `namespace.name` | `arvancloud-dns01-webhook` | Namespace used for webhook resources. |
| `tls.certManager.enabled` | `true` | Create cert-manager self-signed `Issuer` and serving `Certificate`, and inject CA into `APIService`. |
| `rbac.clusterWideSecretAccess` | `true` | Create cluster-wide secret-read RBAC (`get` on `secrets`). |
| `apiService.version` | `v1alpha1` | APIService version; full name is `<version>.<groupName>`. |
