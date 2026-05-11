{{- define "arvancloud-dns01-webhook.name" -}}
{{- default .Chart.Name .Values.service.name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "arvancloud-dns01-webhook.fullname" -}}
{{- if .Values.service.name -}}
{{- .Values.service.name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- .Chart.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{- define "arvancloud-dns01-webhook.namespace" -}}
{{- .Values.namespace.name -}}
{{- end -}}

{{- define "arvancloud-dns01-webhook.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
{{- default (include "arvancloud-dns01-webhook.fullname" .) .Values.serviceAccount.name -}}
{{- else -}}
{{- required "serviceAccount.name is required when serviceAccount.create=false" .Values.serviceAccount.name -}}
{{- end -}}
{{- end -}}

{{- define "arvancloud-dns01-webhook.clusterRoleName" -}}
{{- printf "%s:secrets-reader" (include "arvancloud-dns01-webhook.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "arvancloud-dns01-webhook.apiServiceName" -}}
{{- printf "%s.%s" .Values.apiService.version .Values.groupName -}}
{{- end -}}

{{- define "arvancloud-dns01-webhook.labels" -}}
app.kubernetes.io/name: {{ include "arvancloud-dns01-webhook.fullname" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
helm.sh/chart: {{ printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" }}
{{- end -}}
