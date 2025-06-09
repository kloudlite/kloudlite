{{- $obj := get . "obj" }}
{{- $ownerRefs := get .  "owner-refs" }}
{{- $storageClass := get . "storage-class"}}

{{- $adminPassword := get . "admin-password"  -}}
{{- $adminToken := get . "admin-token"  -}}

{{- with $obj}}
{{- /* gotype: github.com/kloudlite/operator/apis/influxdb.msvc/v1.Service*/ -}}
apiVersion: msvc.kloudlite.io/v1
kind: HelmInfluxDB
metadata:
  name: {{.Name}}
  namespace: {{.Namespace}}
  ownerReferences:  {{ $ownerRefs | toYAML | nindent 4}}
spec:
  fullnameOverride: {{.Name}}
  auth:
    enabled: true
    admin:
      bucket: {{ .Spec.Admin.Bucket }}
      org: {{ .Spec.Admin.Org }}
      password: {{$adminPassword}}
      token: {{$adminToken}}
      username: {{ .Spec.Admin.Username }}

  {{- if .Labels }}
  commonLabels: {{.Labels | toYAML | nindent 4}}
  {{- end}}

  {{- if .Spec.NodeSelector}}
  nodeSelector: {{.Spec.NodeSelector | toYAML | nindent 4}}
  {{- end }}

  metrics:
    enabled: true

  volumePermissions:
    enabled: true

  persistence:
    accessModes:
      - ReadWriteOnce
    enabled: true
    size: {{.Spec.Storage.Size}}
    storageClass: {{$storageClass}}

  influxdb:
    resources:
      requests:
        cpu: {{.Spec.Resources.Cpu.Min}}
        memory: {{.Spec.Resources.Memory}}
      limits:
        cpu: {{.Spec.Resources.Cpu.Max}}
        memory: {{.Spec.Resources.Memory}}
{{- end }}
