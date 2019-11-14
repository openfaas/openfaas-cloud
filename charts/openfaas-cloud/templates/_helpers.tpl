{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "openfaas-cloud.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "openfaas-cloud.fullname" -}}
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
{{- define "openfaas-cloud.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "openfaas-cloud.labels" -}}
app.kubernetes.io/name: {{ include "openfaas-cloud.name" . }}
helm.sh/chart: {{ include "openfaas-cloud.chart" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{- define "openfaas-cloud.tls.dns" -}}
{{- if eq .Values.tls.dnsService "clouddns" }}
          clouddns:
            project: {{ .Values.tls.clouddns.projectID | quote }}
            serviceAccountSecretRef:
              name: "{{ .Values.tls.clouddns.dnsService}}-service-account"
              key: service-account.json
{{- else if eq .Values.tls.dnsService "route53" }}
          route53:
            region: {{ required "A .Values.tls.route53.region is required!" .Values.tls.route53.region }}
  {{- if not .Values.tls.route53.ambientCredentials }}
            accessKeyID: {{ required "A .Values.tls.route53.accessKeyID is required!" .Values.tls.route53.accessKeyID }}
            secretAccessKeySecretRef:
              name: "{{ .Values.tls.dnsService}}-credentials-secret"
              key: secret-access-key
  {{- end }}
{{- else if eq .Values.tls.dnsService "digitalocean" }}
          digitalocean:
            tokenSecretRef:
              name: digitalocean-dns
              key: access-token
{{- end }}
{{- end -}}