{{- $accountName := .AccountName -}}
{{- $name := .Name -}}
{{- $namespace := .Namespace -}}
{{- $labels := .Labels -}}
{{- $annotations := .Annotations -}}
{{- $buildOptions := .BuildOptions -}}
{{- $resource := .Resource -}}
{{- $gitRepo := .GitRepo -}}
{{- $registry := .Registry -}}
{{- $credentialsRef := .CredentialsRef -}}
{{- $caches := .Caches | default list -}}

apiVersion: distribution.kloudlite.io/v1
kind: BuildRun
metadata:
  labels: {{ $labels | toJson }}
  annotations: {{ $annotations | toJson }}
  name: {{ $name }}
  namespace: {{ $namespace }}
spec:
  accountName: {{ $accountName }}
  registry: {{ $registry | toJson }}

  {{- if $buildOptions }}
  buildOptions: {{ $buildOptions | toJson }}
  {{- end }}


  {{- if $caches }}
  caches: {{ $caches | toJson }}
  {{- end }}

  resource: {{ $resource | toJson }}
  gitRepo: {{ $gitRepo | toJson }}
  credentialsRef: {{ $credentialsRef | toJson }}
