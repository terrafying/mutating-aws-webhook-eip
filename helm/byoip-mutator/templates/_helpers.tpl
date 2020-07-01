{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "byoip-mutator.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "byoip-mutator.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "byoip-mutator.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "byoip-mutator.labels" -}}
app.kubernetes.io/name: {{ include "byoip-mutator.name" . }}
helm.sh/chart: {{ include "byoip-mutator.chart" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "byoip-mutator.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
    {{ default (include "byoip-mutator.fullname" .) .Values.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.serviceAccount.name }}
{{- end -}}
{{- end -}}

{{- define "byoip-mutator.selfSignedIssuer" -}}
{{ printf "%s-selfsign" (include "byoip-mutator.fullname" .) }}
{{- end -}}

{{- define "byoip-mutator.rootCAIssuer" -}}
{{ printf "%s-ca" (include "byoip-mutator.fullname" .) }}
{{- end -}}

{{- define "byoip-mutator.rootCACertificate" -}}
{{ printf "%s-ca" (include "byoip-mutator.fullname" .) }}
{{- end -}}

{{- define "byoip-mutator.servingCertificate" -}}
{{ printf "%s-webhook-tls" (include "byoip-mutator.fullname" .) }}
{{- end -}}
