{{- $obj := get . "obj"  -}}
{{- $ownerRefs := get . "owner-refs"  -}}
{{- $elasticUrl := get . "elastic-url"  -}}

{{- $kibanaImageTag := "7.17.3" -}}

{{with $obj}}
{{- /*gotype: operators.kloudlite.io/apis/elasticsearch.msvc/v1.Kibana*/ -}}
apiVersion: msvc.kloudlite.io/v1
kind: HelmKibana
metadata:
  name: {{.Name}}
  namespace: {{.Namespace}}
  ownerReferences: {{ $ownerRefs | toYAML | nindent 4}}
  labels: {{.Labels | toYAML | nindent 4}}
spec:
  affinity: {{include "NodeAffinity" dict | nindent 4 }}
  nodeSelector: {{include "RegionNodeSelector" (dict "region" .Spec.Region) | nindent 4}}
  tolerations: {{include "RegionToleration" (dict "region" .Spec.Region) | nindent 4}}

  elasticsearchHosts: {{$elasticUrl}}

  replicas: {{.Spec.ReplicaCount}}

  fullNameOverride: {{.Name}}-kibana
  imageTag: {{$kibanaImageTag}}
  labels:
    kloudlite.io/region: {{.Spec.Region}}

  ingress:
    enabled: false

  resources:
    limits:
      cpu: {{.Spec.Resources.Cpu.Max}}
      memory: {{.Spec.Resources.Memory}}
    requests:
      cpu: {{.Spec.Resources.Cpu.Min}}
      memory: {{.Spec.Resources.Memory}}
{{end}}
