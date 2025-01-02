{{- $ownerRefs := get . "owner-refs"}}
{{- $storageClass := get . "storage-class" }}
{{- $obj := get . "obj"}}
{{- $elasticPassword := get . "elastic-password" -}}
{{- $labels := get . "labels"  -}}

{{- $priorityClassName := get . "priority-class-name"  | default "stateful" -}}

{{- with $obj}}
{{- /* gotype: github.com/kloudlite/operator/apis/elasticsearch.msvc/v1.Service */ -}}
apiVersion: msvc.kloudlite.io/v1
kind: HelmElasticSearch
metadata:
  name: {{.Name}}
  namespace: {{.Namespace}}
  ownerReferences: {{ $ownerRefs | toYAML | nindent 4}}
  labels: {{$labels | toYAML | nindent 4}}
spec:
  replicas: {{.Spec.ReplicaCount}}
  minimumMasterNodes: {{.Spec.ReplicaCount}}
  fullnameOverride: {{.Name}}

  labels:
    {{ if $labels }}{{$labels | toYAML | nindent 4}}{{end}}
    kloudlite.io/region: {{.Spec.Region}}
    kloudlite.io/stateful-node: "true"

  priorityClassName: {{$priorityClassName}}
  {{ include "NodeAffinity" (dict) | nindent 2 }}
  tolerations: {{ include "RegionToleration" (dict "region" .Spec.Region) | nindent 4 }}
  nodeSelector: {{include "RegionNodeSelector" (dict "region" .Spec.Region) | nindent 4 }}

  volumeClaimTemplate:
    accessModes: [ "ReadWriteOnce" ]
    storageClassName: {{$storageClass}}
    resources:
      requests:
        storage: {{.Spec.Resources.Storage.Size}}

  resources:
    limits:
      cpu: {{.Spec.Resources.Cpu.Max}}
      memory: {{.Spec.Resources.Memory}}
    requests:
      cpu: {{.Spec.Resources.Cpu.Min}}
      memory: {{.Spec.Resources.Memory}}

  secret:
    enabled: true
    password: {{$elasticPassword}}

  service:
    enabled: true
{{- end}}
