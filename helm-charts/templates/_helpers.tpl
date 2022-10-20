{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "grafana-rbac-controller.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "grafana-rbac-controller.fullname" -}}
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
Allow the release namespace to be overridden for multi-namespace deployments in combined charts
*/}}
{{- define "grafana-rbac-controller.namespace" -}}
  {{- if .Values.namespaceOverride -}}
    {{- .Values.namespaceOverride -}}
  {{- else -}}
    {{- .Release.Namespace -}}
  {{- end -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "grafana-rbac-controller.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create default labels
*/}}
{{- define "grafana-rbac-controller.defaultLabels" -}}
{{- $labelChart := include "grafana-rbac-controller.chart" $ -}}
{{- $labelApp := include "grafana-rbac-controller.name" $ -}}
{{- $labels := dict "app" $labelApp "chart" $labelChart "release" .Release.Name "heritage" .Release.Service -}}
{{- $indent := .indent | default 4 -}}
{{ merge .extraLabels $labels | toYaml | indent $indent }}
{{- end -}}

{{/*
Create secret name to store oauth2Proxy client ID, secret and cookies.
*/}}
{{- define "grafana-rbac-controller.oauthCredentials" -}}
  {{- if .Values.oauth2Proxy.credentials.existingSecretName -}}
    {{- .Values.oauth2Proxy.credentials.existingSecretName -}}
  {{- else -}}
    {{- printf "%s-oauth2-proxy-credentials" .Release.Name | trunc 63 | trimSuffix "-" -}}
  {{- end -}}
{{- end -}}

{{/*
Allow the release namespace to be overridden for multi-namespace deployments in combined charts
*/}}
{{- define "grafana-rbac-controller.orgPermissions" -}}
  {{- if .Values.grafanaOrgPermisions.existingConfigMapName -}}
    {{- .Values.grafanaOrgPermisions.existingConfigMapName -}}
  {{- else -}}
    {{- printf "%s-grafana-org-permissions" .Release.Name | trunc 63 | trimSuffix "-" -}}
  {{- end -}}
{{- end -}}

{{/*
Create secret name to store oauth2Proxy client ID, secret and cookies.
*/}}
{{- define "grafana-rbac-controller.grafana-admin-password" -}}
  {{- if .Values.grafanaAdminCredentials.existingSecretName -}}
    {{- .Values.grafanaAdminCredentials.existingSecretName -}}
  {{- else -}}
    {{- printf "%s-grafana-admin-credentials" .Release.Name | trunc 63 | trimSuffix "-" -}}
  {{- end -}}
{{- end -}}

{{/*
Create secret name to store oauth2Proxy client ID, secret and cookies.
*/}}
{{- define "grafana-rbac-controller.google-admin-credentials" -}}
  {{- if .Values.googleAdminCredentials.existingSecretName -}}
    {{- .Values.googleAdminCredentials.existingSecretName -}}
  {{- else -}}
  {{- printf "%s-google-admin-credentials" .Release.Name | trunc 63 | trimSuffix "-" -}}
  {{- end -}}
{{- end -}}

{{/*
Create grafana k8s service name.
*/}}
{{- define "grafana-rbac-controller.existing-grafana-service" -}}
  http://{{- .Values.existingGrafanaService.name -}}.{{ .Values.existingGrafanaService.namespace }}.svc.cluster.local:{{ .Values.existingGrafanaService.portNumber }}
{{- end -}}
