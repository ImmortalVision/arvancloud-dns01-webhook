# cert-manager webhook for ArvanCloud DNS01

Language: English | [فارسی](README.fa.md)

This project provides a cert-manager DNS01 webhook solver for ArvanCloud DNS.

## Background

This repository was created because HTTP-01 validation was not reliably usable for some Iranian-hosted services during US-Iran war-time conditions in Iran. Using DNS-01 through ArvanCloud DNS enables stable certificate issuance and automatic renewal without depending on direct HTTP reachability from ACME validators.

## What this webhook does

- Creates TXT records for ACME DNS01 challenges via ArvanCloud CDN v4 API.
- Deletes matching TXT records during cleanup.
- Reads the ArvanCloud API key from a Kubernetes Secret.

## Verified Arvan API endpoints used

Based on Arvan OpenAPI `cdn-4.0.yml`:

- `POST /domains/{domain}/dns-records` (create TXT)
- `GET /domains/{domain}/dns-records` (list TXT)
- `DELETE /domains/{domain}/dns-records/{id}` (delete)

Base URL default: `https://napi.arvancloud.ir/cdn/4.0`

Authorization header: `Authorization: API KEY ...`

## Local build

```bash
go test ./...
docker build -t ghcr.io/immortalvision/arvancloud-acme-webhook:latest .
```

## Deploy webhook (raw manifests)

1. Install cert-manager in your cluster first.
2. Build and push the webhook image.
3. Update image in `deploy/manifests.yaml`:

   - `ghcr.io/immortalvision/arvancloud-acme-webhook:latest`

4. Apply manifests:

```bash
kubectl apply -f deploy/manifests.yaml
```

5. Create API key secret (example):

```bash
kubectl apply -f deploy/example-secret.yaml
```

6. Create your issuer (example):

```bash
kubectl apply -f deploy/example-clusterissuer.yaml
```

## Solver config

`spec.acme.solvers[].dns01.webhook.config` supports:

- `apiKeySecretRef.name` (required)
- `apiKeySecretRef.key` (required)
- `apiKeySecretRef.namespace` (optional, defaults to challenge namespace)
- `zone` (optional, defaults to cert-manager `resolvedZone`)
- `ttl` (optional, default `120`, must match Arvan allowed TTLs)
- `apiEndpoint` (optional, default `https://napi.arvancloud.ir/cdn/4.0`)

### Example webhook config

```yaml
dns01:
  webhook:
    groupName: acme.arvancloud.ir
    solverName: arvancloud
    config:
      apiKeySecretRef:
        name: arvancloud-api-key
        key: api-key
        namespace: cert-manager
      zone: example.com
      ttl: 120
```

## Verify issuance

After creating a `Certificate` using the issuer:

```bash
kubectl describe challenge -A
kubectl get certificaterequests,certificates -A
```

Expected flow:

- Challenge creates TXT at `_acme-challenge.<domain>` in Arvan DNS.
- Challenge becomes `valid`.
- Certificate reaches `Ready=True`.

## GitHub Actions (CI/CD)

- CI (`.github/workflows/ci.yml`) runs on PRs and pushes to `main`:
  - `go mod tidy` drift check
  - `go test ./...`
  - `docker build`
- Release (`.github/workflows/release.yml`) runs on tag pushes matching `v*` and publishes to GHCR:
  - `ghcr.io/immortalvision/arvancloud-acme-webhook:<tag>`
  - `ghcr.io/immortalvision/arvancloud-acme-webhook:latest` for stable tags (no `-` prerelease suffix)

### Create a release image

```bash
git tag v0.1.0
git push origin v0.1.0
```
