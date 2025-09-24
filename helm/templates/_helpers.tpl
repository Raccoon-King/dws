{{/*
Expand the name of the chart.
*/}}
{{- define "dws.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "dws.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "dws.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "dws.labels" -}}
helm.sh/chart: {{ include "dws.chart" . }}
{{ include "dws.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: document-scanner
{{- with .Values.global.labels }}
{{ toYaml . }}
{{- end }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "dws.selectorLabels" -}}
app.kubernetes.io/name: {{ include "dws.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "dws.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "dws.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the name of the configmap to use
*/}}
{{- define "dws.configMapName" -}}
{{- if .Values.configMap.name }}
{{- .Values.configMap.name }}
{{- else }}
{{- include "dws.fullname" . }}-config
{{- end }}
{{- end }}

{{/*
Create the name of the secret to use
*/}}
{{- define "dws.secretName" -}}
{{- if .Values.secrets.name }}
{{- .Values.secrets.name }}
{{- else }}
{{- include "dws.fullname" . }}-secrets
{{- end }}
{{- end }}

{{/*
Create image reference
*/}}
{{- define "dws.image" -}}
{{- $registry := .Values.image.registry -}}
{{- $repository := .Values.image.repository -}}
{{- $tag := .Values.image.tag | default .Chart.AppVersion -}}
{{- if .Values.ironBank.enabled -}}
{{- $registry = .Values.ironBank.registry -}}
{{- $repository = .Values.ironBank.repository -}}
{{- end -}}
{{- printf "%s/%s:%s" $registry $repository $tag -}}
{{- end }}

{{/*
Create image pull secrets
*/}}
{{- define "dws.imagePullSecrets" -}}
{{- $secrets := list -}}
{{- if .Values.global.imagePullSecrets -}}
{{- $secrets = concat $secrets .Values.global.imagePullSecrets -}}
{{- end -}}
{{- if .Values.image.pullSecrets -}}
{{- $secrets = concat $secrets .Values.image.pullSecrets -}}
{{- end -}}
{{- if $secrets -}}
imagePullSecrets:
{{- range $secrets }}
  - name: {{ . }}
{{- end }}
{{- end -}}
{{- end }}

{{/*
Common annotations
*/}}
{{- define "dws.annotations" -}}
{{- with .Values.global.annotations }}
{{ toYaml . }}
{{- end }}
{{- end }}

{{/*
Pod annotations
*/}}
{{- define "dws.podAnnotations" -}}
checksum/config: {{ include (print $.Template.BasePath "/configmap.yaml") . | sha256sum }}
{{- with .Values.deployment.podAnnotations }}
{{ toYaml . }}
{{- end }}
{{- include "dws.annotations" . }}
{{- end }}

{{/*
Security context for pods (Iron Bank compliance)
*/}}
{{- define "dws.securityContext" -}}
allowPrivilegeEscalation: false
capabilities:
  drop:
    - ALL
readOnlyRootFilesystem: true
runAsNonRoot: true
runAsUser: 65534
runAsGroup: 65534
{{- with .Values.deployment.securityContext }}
{{ toYaml . | nindent 0 }}
{{- end }}
{{- end }}

{{/*
Pod security context (Iron Bank compliance)
*/}}
{{- define "dws.podSecurityContext" -}}
runAsNonRoot: true
runAsUser: 65534
runAsGroup: 65534
fsGroup: 65534
seccompProfile:
  type: RuntimeDefault
{{- with .Values.deployment.podSecurityContext }}
{{ toYaml . | nindent 0 }}
{{- end }}
{{- end }}

{{/*
Environment variables
*/}}
{{- define "dws.env" -}}
{{- range .Values.env }}
- name: {{ .name }}
  value: {{ .value | quote }}
{{- end }}
{{- end }}

{{/*
Volume mounts
*/}}
{{- define "dws.volumeMounts" -}}
{{- range .Values.volumeMounts }}
- name: {{ .name }}
  mountPath: {{ .mountPath }}
  {{- if .readOnly }}
  readOnly: {{ .readOnly }}
  {{- end }}
  {{- if .subPath }}
  subPath: {{ .subPath }}
  {{- end }}
{{- end }}
{{- end }}

{{/*
Volumes
*/}}
{{- define "dws.volumes" -}}
{{- range .Values.volumes }}
- name: {{ .name }}
  {{- if .configMap }}
  configMap:
    name: {{ .configMap.name }}
    {{- if .configMap.defaultMode }}
    defaultMode: {{ .configMap.defaultMode }}
    {{- end }}
  {{- else if .secret }}
  secret:
    secretName: {{ .secret.secretName }}
    {{- if .secret.defaultMode }}
    defaultMode: {{ .secret.defaultMode }}
    {{- end }}
  {{- else if .emptyDir }}
  emptyDir: {{ toYaml .emptyDir | nindent 4 }}
  {{- end }}
{{- end }}
{{- end }}

{{/*
Ingress API version
*/}}
{{- define "dws.ingress.apiVersion" -}}
{{- if .Capabilities.APIVersions.Has "networking.k8s.io/v1/Ingress" -}}
networking.k8s.io/v1
{{- else if .Capabilities.APIVersions.Has "networking.k8s.io/v1beta1/Ingress" -}}
networking.k8s.io/v1beta1
{{- else -}}
extensions/v1beta1
{{- end -}}
{{- end }}

{{/*
HPA API version
*/}}
{{- define "dws.hpa.apiVersion" -}}
{{- if .Capabilities.APIVersions.Has "autoscaling/v2/HorizontalPodAutoscaler" -}}
autoscaling/v2
{{- else if .Capabilities.APIVersions.Has "autoscaling/v2beta2/HorizontalPodAutoscaler" -}}
autoscaling/v2beta2
{{- else -}}
autoscaling/v2beta1
{{- end -}}
{{- end }}

{{/*
Network Policy API version
*/}}
{{- define "dws.networkPolicy.apiVersion" -}}
{{- if .Capabilities.APIVersions.Has "networking.k8s.io/v1/NetworkPolicy" -}}
networking.k8s.io/v1
{{- else -}}
networking.k8s.io/v1beta1
{{- end -}}
{{- end }}