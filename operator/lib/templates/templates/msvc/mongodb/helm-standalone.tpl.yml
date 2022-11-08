{{- $obj :=  get . "object" }}
{{- $ownerRefs := get . "owner-refs"}}
{{- $storageClass := get . "storage-class"}}
{{- $freeze := get . "freeze" | default false}}
{{- $priorityClassName := get . "priority-class-name"  | default "stateful" -}}
{{- $existingSecret := get . "existing-secret" -}}

{{- with $obj }}
{{- /* gotype: operators.kloudlite.io/apis/mongodb.msvc/v1.StandaloneService*/ -}}
{{$labels := .Labels | default dict}}
apiVersion: msvc.kloudlite.io/v1
kind: HelmMongoDB
metadata:
  name: {{.Name}}
  namespace: {{.Namespace}}
  labels: {{ $labels | toYAML | nindent 4 }}
  ownerReferences: {{ $ownerRefs | toYAML | nindent 4 }}
spec:
  global:
    storageClass: {{$storageClass}}
  fullnameOverride: {{.Name}}
  image:
    tag: 5.0.8-debian-10-r20

  architecture: standalone
  useStatefulSet: true

  replicaCount: {{ if $freeze}}0{{else}}{{.Spec.ReplicaCount}}{{end}}

  commonLabels: {{$labels | toYAML | nindent 4}}

  podLabels:
    {{$labels | toYAML | nindent 4}}
    kloudlite.io/region: {{.Spec.Region}}
    kloudlite.io/stateful-node: "true"

  auth:
    enabled: true
    existingSecret: {{$existingSecret}}

  persistence:
    enabled: true
    size: {{.Spec.Resources.Storage.Size}}

  volumePermissions:
    enabled: true

  metrics:
    enabled: true

  livenessProbe:
    enabled: true
    timeoutSeconds: 15

  readinessProbe:
    enabled: true
    initialDelaySeconds: 10
    periodSeconds: 30
    timeoutSeconds: 20

  priorityClassName: {{$priorityClassName}}
  affinity: {{include "NodeAffinity" (dict) | nindent 4 }}
  tolerations:
    {{include "RegionToleration" (dict "region" .Spec.Region) | nindent 4}}
    {{ if .Spec.Tolerations }}
    {{.Spec.Tolerations | toYAML | nindent 4}}
    {{ end }}
  nodeSelector:
    {{ include "RegionNodeSelector" (dict "region" .Spec.Region) | nindent 4 }}
    {{ if .Spec.NodeSelector }}
    {{.Spec.NodeSelector | toYAML | nindent 4}}
    {{ end }}

  resources:
    requests:
      cpu: {{.Spec.Resources.Cpu.Min}}
      memory: {{ .Spec.Resources.Memory }}
    limits:
      cpu: {{.Spec.Resources.Cpu.Max}}
      memory: {{.Spec.Resources.Memory}}
{{- end }}
