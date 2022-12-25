{{- $name := get . "name" -}}
{{- $namespace := get . "namespace" -}}
{{- $ownerRefs := get . "owner-refs" | default list -}}
{{- $labels := get . "labels" |default dict -}}
{{- $stringData := get . "string-data" -}}
{{- $secretType := get . "secret-type" | default "Opaque" -}}
{{- $data := get . "data" | default dict -}}

apiVersion: crds.kloudlite.io/v1
kind: Secret
metadata:
  name: {{$name}}
  namespace: {{$namespace}}
  ownerReferences: {{$ownerRefs | toYAML | nindent 4}}
  labels: {{$labels | toYAML | nindent 4}}
  finalizers:
    - foregroundDeletion
type: {{$secretType}}
{{/*stringData: {{$stringData | toYAML | nindent 2}}*/}}
stringData:
{{- range $k, $v := mustFromJson ($stringData | toJson) -}}
{{$k | nindent 2}}: {{$v | squote}}
{{end}}
data: {{$data}}
