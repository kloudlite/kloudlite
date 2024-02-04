{{- $obj := get . "obj"  -}}
{{- $ownerRefs := get . "owner-refs" -}}
{{- $storageClass := get . "storage-class"}}
{{- $rootPassword := get . "root-password"  -}}
{{- $priorityClassName := get . "priority-class-name"  | default "stateful" -}}
{{- $labels := get . "labels"  -}}


{{- with $obj}}
{{- /*gotype: github.com/kloudlite/operator/apis/neo4j.msvc/v1.StandaloneService*/ -}}
apiVersion: msvc.kloudlite.io/v1
kind: HelmNeo4jStandalone
metadata:
  name: {{.Name}}
  namespace: {{.Namespace}}
  ownerReferences: {{$ownerRefs| toYAML | nindent 4}}
  labels:
    {{if $labels}}{{$labels | toYAML | nindent 4}}{{end}}
    kloudlite.io/region: {{.Spec.Region}}
    kloudlite.io/stateful-node: "true"
spec:
  neo4j:
    name: {{.Name}}
    resources:
      cpu: {{.Spec.Resources.Cpu.Max}}
      memory: "{{.Spec.Resources.Memory}}"

    password: "{{$rootPassword}}"

    labels:
      kloudlite.io/region: {{.Spec.Region}}
      kloudlite.io/stateful-node: "true"

    # Uncomment to use enterprise edition
    #edition: "enterprise"
    #acceptLicenseAgreement: "yes"

  volumes:
    data:
      mode: "dynamic"
      dynamic:
        storageClassName: "{{$storageClass}}"
        accessModes:
          - ReadWriteOnce
        requests:
          storage: {{.Spec.Resources.Storage.Size}}

  podSpec:
    tolerations:
      {{include "RegionToleration" (dict "region" .Spec.Region) | nindent 8 }}
      {{if .Spec.Tolerations}}{{.Spec.Tolerations | toYAML | nindent 8}}{{end}}

    nodeSelector:
      {{include "RegionNodeSelector" (dict "region" .Spec.Region) | nindent 8}}
      {{if .Spec.NodeSelector}}{{.Spec.NodeSelector | toYAML |nindent 8}}{{end}}

    labels:
      kloudlite.io/region: {{.Spec.Region}}
      kloudlite.io/stateful-node: "true"

    {{include "NodeAffinity" dict | nindent 4}}
    priorityClassName: {{$priorityClassName}}

  statefulset:
    metadata:
      labels:
        kloudlite.io/region: {{.Spec.Region}}
        kloudlite.io/stateful-node: "true"

  services:
    neo4j:
      enabled: false
      spec:
        type: ClusterIP
{{- end }}
