# چارت Helm وبهوک arvancloud-dns01-webhook

این چارت برای دیپلوی وبهوک DNS01 آروان‌کلاد (سازگار با cert-manager) است.

## پیش‌نیازها

- کلاستر Kubernetes
- نصب cert-manager

## نصب

```bash
helm upgrade --install arvancloud-dns01-webhook ./charts/arvancloud-dns01-webhook
```

## نمونه override

```bash
helm upgrade --install arvancloud-dns01-webhook ./charts/arvancloud-dns01-webhook \
  --set image.tag=v0.1.0 \
  --set groupName=acme.arvancloud.ir
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
| `rbac.clusterWideSecretAccess` | `true` | ایجاد RBAC سراسری برای خواندن secret (`get` روی `secrets`) |
| `apiService.version` | `v1alpha1` | نسخه APIService؛ نام کامل: `<version>.<groupName>` |
