{{- $obj := get . "object"}}
{{- $ownerRefs := get . "owner-refs" }}
{{- $storageClass := get . "storage-class"}}
{{/*{{- $aclAccountsMap := get . "acl-account-map"}}*/}}

{{/*{{- $freeze := get . "freeze" | default false}}*/}}
{{- $aclConfigmapName := get . "acl-configmap-name"  -}}
{{- $rootPassword := get . "root-password"  -}}
{{- $priorityClassName := get . "priority-class-name"  | default "stateful" -}}

{{- with $obj }}
{{- /*gotype: github.com/kloudlite/operator/apis/redis-standalone.msvc/v1.Service */ -}}
apiVersion: msvc.kloudlite.io/v1
kind: HelmRedis
metadata:
  name: {{.Name}}
  namespace: {{.Namespace}}
  ownerReferences: {{ $ownerRefs | toYAML | nindent 4 }}
  {{- if .Labels }}
  labels: {{ .Labels | toYAML | nindent 4 }}
  {{- end}}
spec:
  global:
   storageClass: {{$storageClass}}
  fullnameOverride: {{.Name}}
  image:
    tag: 6.2.7-debian-10-r0

  {{- if .Labels }}
  commonLabels: {{ .Labels | toYAML | nindent 4 }}
  {{- end}}

  {{- if .Spec.NodeSelector}}
  nodeSelector: {{.Spec.NodeSelector | toYAML | nindent 4}}
  {{- end}}

  auth:
    enabled: true
    password: {{$rootPassword}}

  replica:
    replicaCount: 0

  architecture: standalone

  existingConfigmap: {{$aclConfigmapName}}
{{/*  commonConfiguration: |+*/}}
{{/*    {{- if $aclAccountsMap}}*/}}
{{/*    {{- range $k, $v := $aclAccountsMap}}*/}}
{{/*    {{ $v }}*/}}
{{/*    {{- end }}*/}}
{{/*    {{- end }}*/}}

  master:
    count: {{.Spec.ReplicaCount}}
    resources:
      requests:
        cpu: {{ .Spec.Resources.Cpu.Min}}
        memory: {{.Spec.Resources.Memory}}
      limits:
        cpu: {{ .Spec.Resources.Cpu.Max}}
        memory: {{.Spec.Resources.Memory}}
    persistence:
      enabled: true
      size: {{.Spec.Resources.Storage.Size}}

    podLabels:
      {{ if .Labels}}{{.Labels | toYAML | nindent 6 }}{{ end}}
      {{- if .Spec.Region }}
      kloudlite.io/region: {{.Spec.Region}}
      {{- end }}
      kloudlite.io/stateful-node: "true"

    priorityClassName: {{$priorityClassName}}
    affinity: {{include "NodeAffinity" (dict) | toYAML | nindent 6}}
    {{- if .Spec.Region }}
    tolerations: {{include "RegionToleration" (dict "region" .Spec.Region) | nindent 6}}
    nodeSelector: {{include "RegionNodeSelector" (dict "region" .Spec.Region) | nindent 6}}
    {{- end }}

  replica:
    {{- if .Labels}}
    podLabels: {{.Labels | toYAML | nindent 6 }}
    {{- end}}

    priorityClassName: {{$priorityClassName}}
    affinity: {{include "NodeAffinity" (dict) | toYAML | nindent 6}}
    {{- if .Spec.Region}}
    tolerations: {{include "RegionToleration" (dict "region" .Spec.Region) | nindent 6}}
    nodeSelector: {{include "RegionNodeSelector" (dict "region" .Spec.Region) | nindent 6}}
    {{- end }}

  volumePermissions:
    enabled: true

  metrics:
    enabled: true
{{- end}}
