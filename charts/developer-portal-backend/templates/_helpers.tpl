{{/*
Expand the name of the chart.
*/}}
{{- define "developer-portal-backend.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "developer-portal-backend.fullname" -}}
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
{{- define "developer-portal-backend.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "developer-portal-backend.labels" -}}
helm.sh/chart: {{ include "developer-portal-backend.chart" . }}
{{ include "developer-portal-backend.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "developer-portal-backend.selectorLabels" -}}
app.kubernetes.io/name: {{ include "developer-portal-backend.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "developer-portal-backend.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "developer-portal-backend.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Database host
*/}}
{{- define "developer-portal-backend.databaseHost" -}}
{{- if .Values.database.useExternal }}
{{- .Values.database.external.host }}
{{- else }}
{{- printf "%s-postgresql" .Release.Name }}
{{- end }}
{{- end }}

{{/*
Database port
*/}}
{{- define "developer-portal-backend.databasePort" -}}
{{- if .Values.database.useExternal }}
{{- .Values.database.external.port }}
{{- else }}
{{- "5432" }}
{{- end }}
{{- end }}

{{/*
Database name
*/}}
{{- define "developer-portal-backend.databaseName" -}}
{{- if .Values.database.useExternal }}
{{- .Values.database.external.name }}
{{- else }}
{{- .Values.postgresql.auth.database }}
{{- end }}
{{- end }}

{{/*
Database user
*/}}
{{- define "developer-portal-backend.databaseUser" -}}
{{- if .Values.database.useExternal }}
{{- .Values.database.external.user }}
{{- else }}
{{- .Values.postgresql.auth.username }}
{{- end }}
{{- end }}

{{/*
Database password
*/}}
{{- define "developer-portal-backend.databasePassword" -}}
{{- if .Values.database.useExternal }}
{{- .Values.database.external.password }}
{{- else }}
{{- .Values.postgresql.auth.password }}
{{- end }}
{{- end }}


