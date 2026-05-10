# AGENTS

## Project Scope
- This is a cert-manager DNS01 webhook implementation for ArvanCloud DNS.
- Runtime group name is set by `GROUP_NAME` env var (deployment uses `acme.arvancloud.ir`).

## Code Layout
- `main.go`: webhook server bootstrap using cert-manager webhook framework.
- `pkg/arvancloud/solver.go`: cert-manager solver implementation (`Present`, `CleanUp`, config parsing, secret loading).
- `pkg/arvancloud/client.go`: Arvan DNS API client.
- `pkg/arvancloud/types.go`: solver config schema/defaults/validation.
- `deploy/`: raw Kubernetes manifests and usage examples.

## Verified Commands
- Compile/check: `go test ./...`
- Refresh deps after changing imports: `go mod tidy`
- Build container: `docker build -t ghcr.io/immortalvision/arvancloud-dns01-webhook:latest .`

## Arvan API Contracts Used (from OpenAPI)
- Base URL default: `https://napi.arvancloud.ir/cdn/4.0`
- Create TXT: `POST /domains/{domain}/dns-records`
- List records: `GET /domains/{domain}/dns-records?type=txt`
- Delete record: `DELETE /domains/{domain}/dns-records/{id}`
- Auth header uses `Authorization`; implementation expects key in `API KEY ...` format (auto-prefixes when missing).
- TXT record payload uses `value.text`; TTL must be one of Arvan allowed enum values.

## Webhook Config Conventions
- Required config fields: `apiKeySecretRef.name`, `apiKeySecretRef.key`.
- Optional: `apiKeySecretRef.namespace` (defaults to challenge namespace), `zone`, `ttl`, `apiEndpoint`.
- If `zone` is omitted, solver uses cert-manager `resolvedZone`; set `zone` explicitly only when you need strict zone matching.

## Operational Gotchas
- RBAC in `deploy/manifests.yaml` grants cluster-wide secret `get`; tighten scope if you standardize secret location.
- `deploy/manifests.yaml` defaults to `ghcr.io/immortalvision/arvancloud-dns01-webhook:latest`; pin explicit version tags in production.
- `deploy/examples/common/secret.yaml` contains sample credentials only; never commit real API keys.

## CI/CD
- CI workflow: `.github/workflows/ci.yml` runs `go mod tidy` drift check, `go test ./...`, and `docker build` on PRs and `main` pushes.
- Release workflow: `.github/workflows/release.yml` publishes GHCR image on `v*` tags.
- `latest` is published only for non-prerelease tags (tags without `-`, e.g. `v1.2.3`).
