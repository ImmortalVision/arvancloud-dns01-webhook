# چارت Helm وبهوک arvancloud-dns01-webhook

این چارت برای دیپلوی وبهوک DNS01 آروان‌کلاد (سازگار با cert-manager) است.

## پیش‌نیازها

- کلاستر Kubernetes
- نصب cert-manager

## نصب

```bash
helm upgrade --install arvancloud-dns01-webhook ./charts/arvancloud-dns01-webhook \
  -n arvancloud-dns01-webhook
```

## نمونه override

```bash
helm upgrade --install arvancloud-dns01-webhook ./charts/arvancloud-dns01-webhook \
  -n arvancloud-dns01-webhook \
  --set image.tag=v0.1.0 \
  --set groupName=acme.arvancloud.ir
```

## ClusterIssuer پروداکشن (به‌صورت پیش‌فرض فعال)

این چارت به‌صورت پیش‌فرض یک `ClusterIssuer` پروداکشن ACME می‌سازد.
مقدار `clusterIssuer.email` را تنظیم کنید و `clusterIssuer.solver.apiKeySecretRef` را به secret حاوی API key آروان‌کلاد اشاره دهید.

```bash
helm upgrade --install arvancloud-dns01-webhook ./charts/arvancloud-dns01-webhook \
  -n arvancloud-dns01-webhook \
  --set clusterIssuer.email=you@example.com \
  --set clusterIssuer.solver.apiKeySecretRef.name=arvancloud-api-key \
  --set clusterIssuer.solver.apiKeySecretRef.key=api-key \
  --set clusterIssuer.solver.apiKeySecretRef.namespace=cert-manager
```

اگر `clusterIssuer.enabled=true` باشد و مقادیر لازم تکمیل نشده باشند، Helm با پیام خطای واضح fail می‌شود.

## ClusterIssuer استیجینگ اختیاری

برای تست امن‌تر قبل از صدور پروداکشن، می‌توانید یک `ClusterIssuer` استیجینگ هم بسازید.

```bash
helm upgrade --install arvancloud-dns01-webhook ./charts/arvancloud-dns01-webhook \
  -n arvancloud-dns01-webhook \
  --set clusterIssuer.email=you@example.com \
  --set clusterIssuerStaging.enabled=true \
  --set clusterIssuerStaging.email=you@example.com
```

## حذف

```bash
helm uninstall arvancloud-dns01-webhook -n arvancloud-dns01-webhook
```

## مقادیر مهم

| مقدار | پیش‌فرض | توضیح |
| --- | --- | --- |
| `groupName` | `acme.arvancloud.ir` | گروه API وبهوک و متغیر `GROUP_NAME` برای تطابق solver در cert-manager |
| `image.repository` | `ghcr.io/immortalvision/arvancloud-dns01-webhook` | ریپازیتوری ایمیج وبهوک |
| `image.tag` | `latest` | تگ ایمیج وبهوک |
| `namespace.create` | `true` | ایجاد namespace هدف توسط چارت |
| `namespace.name` | `arvancloud-dns01-webhook` | namespace منابع وبهوک |
| `tls.certManager.enabled` | `true` | ساخت `Issuer` و `Certificate` توسط cert-manager و تزریق CA در `APIService` |
| `clusterIssuer.enabled` | `true` | ساخت `ClusterIssuer` پروداکشن ACME برای DNS01 با این وبهوک |
| `clusterIssuer.name` | `letsencrypt-arvancloud` | نام `ClusterIssuer` ساخته‌شده |
| `clusterIssuer.email` | `""` | ایمیل حساب ACME (در صورت `clusterIssuer.enabled=true` الزامی است) |
| `clusterIssuer.server` | `https://acme-v02.api.letsencrypt.org/directory` | آدرس ACME server (به‌صورت پیش‌فرض پروداکشن) |
| `clusterIssuer.privateKeySecretName` | `letsencrypt-arvancloud-account-key` | نام secret کلید خصوصی حساب ACME |
| `clusterIssuer.solver.solverName` | `arvancloud` | نام solver وبهوک ثبت‌شده توسط این پروژه |
| `clusterIssuer.solver.apiKeySecretRef.name` | `arvancloud-api-key` | نام secret حاوی API key آروان‌کلاد |
| `clusterIssuer.solver.apiKeySecretRef.key` | `api-key` | کلید داخل secret برای مقدار API key |
| `clusterIssuer.solver.apiKeySecretRef.namespace` | `cert-manager` | namespace مربوط به secret API key |
| `clusterIssuer.solver.zone` | `""` | ناحیه اختیاری برای strict matching |
| `clusterIssuer.solver.ttl` | `120` | مقدار TTL رکورد TXT ارسالی به API آروان‌کلاد |
| `clusterIssuer.solver.apiEndpoint` | `https://napi.arvancloud.ir/cdn/4.0` | آدرس پایه API DNS آروان‌کلاد |
| `clusterIssuerStaging.enabled` | `false` | ساخت یک `ClusterIssuer` استیجینگ اضافی |
| `clusterIssuerStaging.name` | `letsencrypt-staging-arvancloud` | نام `ClusterIssuer` استیجینگ ساخته‌شده |
| `clusterIssuerStaging.email` | `""` | ایمیل حساب ACME برای استیجینگ (در صورت فعال بودن الزامی است) |
| `clusterIssuerStaging.server` | `https://acme-staging-v02.api.letsencrypt.org/directory` | آدرس ACME server استیجینگ |
| `clusterIssuerStaging.privateKeySecretName` | `letsencrypt-staging-arvancloud-account-key` | نام secret کلید خصوصی حساب ACME استیجینگ |
| `clusterIssuerStaging.solver.solverName` | `arvancloud` | نام solver وبهوک برای issuer استیجینگ |
| `clusterIssuerStaging.solver.apiKeySecretRef.name` | `arvancloud-api-key` | نام secret حاوی API key برای issuer استیجینگ |
| `clusterIssuerStaging.solver.apiKeySecretRef.key` | `api-key` | کلید داخل secret برای مقدار API key در issuer استیجینگ |
| `clusterIssuerStaging.solver.apiKeySecretRef.namespace` | `cert-manager` | namespace مربوط به secret API key استیجینگ |
| `clusterIssuerStaging.solver.zone` | `""` | ناحیه اختیاری برای strict matching در issuer استیجینگ |
| `clusterIssuerStaging.solver.ttl` | `120` | مقدار TTL رکورد TXT ارسالی به API آروان‌کلاد برای issuer استیجینگ |
| `clusterIssuerStaging.solver.apiEndpoint` | `https://napi.arvancloud.ir/cdn/4.0` | آدرس پایه API DNS آروان‌کلاد برای issuer استیجینگ |
| `rbac.clusterWideSecretAccess` | `true` | ایجاد RBAC سراسری برای خواندن secret (`get` روی `secrets`) |
| `apiService.version` | `v1alpha1` | نسخه APIService؛ نام کامل: `<version>.<groupName>` |
