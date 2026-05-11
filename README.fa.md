# وبهوک cert-manager برای DNS01 آروان‌کلاد

زبان: فارسی | [English](README.md)

این پروژه یک حل‌کننده DNS01 برای cert-manager روی DNS آروان‌کلاد فراهم می‌کند.

## پیش‌زمینه

این ریپازیتوری به این دلیل ساخته شد که در زمان جنگ ایران و آمریکا، برای بعضی سرویس‌های میزبانی‌شده در ایران استفاده از اعتبارسنجی HTTP-01 قابل اتکا نبود. استفاده از DNS-01 روی DNS آروان‌کلاد باعث می‌شود صدور و تمدید خودکار گواهی، بدون وابستگی به دسترسی مستقیم HTTP از سمت اعتبارسنج ACME، پایدارتر انجام شود.

## کارکرد این وبهوک

- ایجاد رکورد TXT برای چالش DNS01 از طریق API نسخه ۴ آروان‌کلاد
- حذف رکورد TXT متناظر در مرحله پاک‌سازی
- خواندن API Key آروان‌کلاد از Kubernetes Secret

## APIهای استفاده‌شده آروان‌کلاد

- `POST /domains/{domain}/dns-records` (ایجاد TXT)
- `GET /domains/{domain}/dns-records` (لیست TXT)
- `DELETE /domains/{domain}/dns-records/{id}` (حذف رکورد)

آدرس پیش‌فرض API: `https://napi.arvancloud.ir/cdn/4.0`

هدر احراز هویت: `Authorization: API KEY ...`

## بیلد محلی

```bash
go test ./...
docker build -t ghcr.io/immortalvision/arvancloud-dns01-webhook:latest .
```

## دیپلوی وبهوک (مانیفست خام)

1. ابتدا cert-manager را روی کلاستر نصب کنید.
2. ایمیج وبهوک را بیلد و پوش کنید.
3. در `deploy/manifests.yaml` تگ ایمیج را به نسخه مناسب تغییر دهید.
4. مانیفست‌ها را اعمال کنید:

```bash
kubectl apply -f deploy/manifests.yaml
```

5. سکرت API Key نمونه را اعمال کنید:

```bash
kubectl apply -f deploy/examples/common/secret.yaml
```

6. ClusterIssuer لتس‌انکریپت را اعمال کنید:

- استیجینگ (برای تست اولیه): `deploy/examples/staging/clusterissuer.yaml`
- پروداکشن: `deploy/examples/production/clusterissuer.yaml`

```bash
kubectl apply -f deploy/examples/staging/clusterissuer.yaml
```

## تنظیمات Solver

در `spec.acme.solvers[].dns01.webhook.config`:

- `apiKeySecretRef.name` (اجباری)
- `apiKeySecretRef.key` (اجباری)
- `apiKeySecretRef.namespace` (اختیاری، پیش‌فرض namespace چالش)
- `zone` (اختیاری، پیش‌فرض `resolvedZone` از cert-manager)
- `ttl` (اختیاری، پیش‌فرض `120`)
- `apiEndpoint` (اختیاری، پیش‌فرض `https://napi.arvancloud.ir/cdn/4.0`)

برای جلوگیری از خطای عدم تطابق زون، فقط زمانی `zone` را تنظیم کنید که واقعا به strict matching نیاز دارید.

## مثال گواهی استیجینگ

```bash
kubectl apply -f deploy/examples/staging/certificate.yaml
```

برای بررسی وضعیت:

```bash
kubectl get challenge -A
kubectl describe challenge -A
kubectl get certificate,certificaterequest -A
```

## CI/CD

- CI: در `.github/workflows/ci.yml` روی PRها و push به `main`
- Release: در `.github/workflows/release.yml` روی تگ‌های `v*` و انتشار روی GHCR
