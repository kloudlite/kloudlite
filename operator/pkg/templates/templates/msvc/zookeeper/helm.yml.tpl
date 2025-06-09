{{- $obj := get . "obj" -}}
{{- $ownerRefs := get . "owner-refs" -}}
{{- $storageClass := get . "storage-class" -}}
{{- $rootPassword := get . "root-password" -}}
{{- $helmSecret := get . "helm-secret" -}}

{{- with $obj }}
{{- /* gotype: github.com/kloudlite/operator/apis/zookeeper.msvc/v1.Service */ -}}
apiVersion: msvc.kloudlite.io/v1
kind: HelmZookeeper
metadata:
  name: {{.Name}}
  namespace: {{.Namespace}}
  labels: {{.Labels | toYAML | nindent 4}}
  ownerReferences: {{$ownerRefs | toYAML | nindent 4}}
spec:
  global:
    storageClass: {{$storageClass}}
  fullNameOverride: {{$obj.Name}}
  commonLabels: {{.Labels | toYAML | nindent 4}}

  image:
    tag: 3.8.0-debian-11-r36

  replicaCount: {{.Spec.ReplicaCount}}
  resources:
    limits:
      memory: {{.Spec.Resources.Memory}}
      cpu: {{.Spec.Resources.Cpu.Max}}
    requests:
      memory: {{.Spec.Resources.Memory}}
      cpu: {{.Spec.Resources.Cpu.Min}}

  podLabels: {{.Labels | toYAML | nindent 4}}
  {{- if .Spec.NodeSelector }}
  nodeSelector: {{.Spec.NodeSelector | toYAML | nindent 4}}
  {{- end}}

  persistence:
    enabled: true
    accessModes:
      - ReadWriteOnce
    size: {{.Spec.Resources.Storage.Size}}

  auth:
    client:
      enabled: false
{{/*      existingSecret: {{$helmSecret}}*/}}
{{/*      clientUser: "root"*/}}
{{/*      clientPassword: "{{$rootPassword}}"*/}}
{{/*      serverUsers: "root"*/}}
{{/*      serverPasswords: "{{$rootPassword}}"*/}}
{{- end}}
