{{- $obj := get . "obj"  -}}
{{- $ownerRefs := get . "owner-refs" -}}
{{- $storageClass := get . "storage-class"}}
{{- $rootPassword := get . "root-password"  -}}

{{- with $obj}}
{{- /*gotype: operators.kloudlite.io/apis/neo4j.msvc/v1.StandaloneService*/ -}}
apiVersion: msvc.kloudlite.io/v1
kind: HelmNeo4jStandalone
metadata:
  name: {{.Name}}
  namespace: {{.Namespace}}
  ownerReferences: {{$ownerRefs| toYAML | nindent 4}}
spec:
  neo4j:
    spec:
      type: ClusterIP
    resources:
      cpu: {{.Spec.Resources.Cpu.Max}}
      memory: "{{.Spec.Resources.Memory}}"

    password: "{{$rootPassword}}"

    # Uncomment to use enterprise edition
    #edition: "enterprise"
    #acceptLicenseAgreement: "yes"

  {{- if .Spec.NodeSelector }}
  nodeSelector: {{.Spec.NodeSelector | toYAML | nindent 4}}
  {{- end }}

  volumes:
    data:
      mode: "dynamic"
      dynamic:
        storageClassName: "{{$storageClass}}"
        accessModes:
          - ReadWriteOnce
        requests:
          storage: {{.Spec.Storage.Size}}
      volume:
        gcePersistentDisk:
          pdName: "my-neo4j-disk"

  podSpec:
    {{- if .Spec.Tolerations }}
    tolerations: {{.Spec.Tolerations}}
    {{- end }}
{{- end }}
