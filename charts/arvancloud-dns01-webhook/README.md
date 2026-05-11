# arvancloud-dns01-webhook Helm Chart

Helm chart for deploying the cert-manager DNS01 webhook for ArvanCloud.

## Prerequisites

- Kubernetes cluster
- cert-manager installed

## Install

```bash
helm upgrade --install arvancloud-dns01-webhook ./charts/arvancloud-dns01-webhook \
  -n arvancloud-dns01-webhook
```

## Example override

```bash
helm upgrade --install arvancloud-dns01-webhook ./charts/arvancloud-dns01-webhook \
  -n arvancloud-dns01-webhook \
  --set image.tag=v0.1.0 \
  --set groupName=acme.arvancloud.ir
```

## Production ClusterIssuer (enabled by default)

This chart creates a production ACME `ClusterIssuer` by default.
Set `clusterIssuer.email` and point `clusterIssuer.solver.apiKeySecretRef` to your ArvanCloud API key secret.

```bash
helm upgrade --install arvancloud-dns01-webhook ./charts/arvancloud-dns01-webhook \
  -n arvancloud-dns01-webhook \
  --set clusterIssuer.email=you@example.com \
  --set clusterIssuer.solver.apiKeySecretRef.name=arvancloud-api-key \
  --set clusterIssuer.solver.apiKeySecretRef.key=api-key \
  --set clusterIssuer.solver.apiKeySecretRef.namespace=cert-manager
```

If `clusterIssuer.enabled=true` and required values are missing, Helm fails fast with a clear error.

## Optional staging ClusterIssuer

You can additionally create a staging `ClusterIssuer` for safer testing before production issuance.

```bash
helm upgrade --install arvancloud-dns01-webhook ./charts/arvancloud-dns01-webhook \
  -n arvancloud-dns01-webhook \
  --set clusterIssuer.email=you@example.com \
  --set clusterIssuerStaging.enabled=true \
  --set clusterIssuerStaging.email=you@example.com
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
| `clusterIssuer.enabled` | `true` | Create a production ACME `ClusterIssuer` for DNS01 via this webhook. |
| `clusterIssuer.name` | `letsencrypt-arvancloud` | Name of the generated `ClusterIssuer`. |
| `clusterIssuer.email` | `""` | ACME account email (required when `clusterIssuer.enabled=true`). |
| `clusterIssuer.server` | `https://acme-v02.api.letsencrypt.org/directory` | ACME server URL (production by default). |
| `clusterIssuer.privateKeySecretName` | `letsencrypt-arvancloud-account-key` | Secret name for ACME account private key. |
| `clusterIssuer.solver.solverName` | `arvancloud` | Webhook solver name registered by this project. |
| `clusterIssuer.solver.apiKeySecretRef.name` | `arvancloud-api-key` | Secret name containing ArvanCloud API key. |
| `clusterIssuer.solver.apiKeySecretRef.key` | `api-key` | Secret key containing ArvanCloud API key value. |
| `clusterIssuer.solver.apiKeySecretRef.namespace` | `cert-manager` | Namespace where the API key secret exists. |
| `clusterIssuer.solver.zone` | `""` | Optional explicit zone for strict matching. |
| `clusterIssuer.solver.ttl` | `120` | TXT record TTL sent to ArvanCloud API. |
| `clusterIssuer.solver.apiEndpoint` | `https://napi.arvancloud.ir/cdn/4.0` | ArvanCloud DNS API base endpoint. |
| `clusterIssuerStaging.enabled` | `false` | Create an additional staging ACME `ClusterIssuer`. |
| `clusterIssuerStaging.name` | `letsencrypt-staging-arvancloud` | Name of the generated staging `ClusterIssuer`. |
| `clusterIssuerStaging.email` | `""` | ACME account email for staging (required when enabled). |
| `clusterIssuerStaging.server` | `https://acme-staging-v02.api.letsencrypt.org/directory` | Staging ACME server URL. |
| `clusterIssuerStaging.privateKeySecretName` | `letsencrypt-staging-arvancloud-account-key` | Secret name for staging ACME account private key. |
| `clusterIssuerStaging.solver.solverName` | `arvancloud` | Webhook solver name for staging issuer. |
| `clusterIssuerStaging.solver.apiKeySecretRef.name` | `arvancloud-api-key` | Secret name containing ArvanCloud API key for staging issuer. |
| `clusterIssuerStaging.solver.apiKeySecretRef.key` | `api-key` | Secret key containing ArvanCloud API key value for staging issuer. |
| `clusterIssuerStaging.solver.apiKeySecretRef.namespace` | `cert-manager` | Namespace where staging API key secret exists. |
| `clusterIssuerStaging.solver.zone` | `""` | Optional explicit zone for strict matching in staging issuer. |
| `clusterIssuerStaging.solver.ttl` | `120` | TXT record TTL sent to ArvanCloud API for staging issuer. |
| `clusterIssuerStaging.solver.apiEndpoint` | `https://napi.arvancloud.ir/cdn/4.0` | ArvanCloud DNS API base endpoint for staging issuer. |
| `rbac.clusterWideSecretAccess` | `true` | Create cluster-wide secret-read RBAC (`get` on `secrets`). |
| `apiService.version` | `v1alpha1` | APIService version; full name is `<version>.<groupName>`. |
