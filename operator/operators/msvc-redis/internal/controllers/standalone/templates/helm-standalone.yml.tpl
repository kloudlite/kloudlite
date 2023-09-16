{{- $name := get . "name" }}
{{- $namespace := get . "namespace" }}
{{- $labels := get . "labels" | default dict }}
{{- $ownerRefs := get . "owner-refs" | default list }}

{{- $storageSize := get . "storage-size" }}
{{- $storageClass := get . "storage-class" }}

{{- $nodeSelector := get . "node-selector" | default dict }}

{{- $requestsCpu := get . "requests-cpu" }}
{{- $requestsMem := get . "requests-mem" }}
{{- $limitsCpu := get . "limits-cpu" }}
{{- $limitsMem := get . "limits-mem" }}

{{- $obj := get . "object"}}
{{/*{{- $ownerRefs := get . "owner-refs" }}*/}}
{{/*{{- $storageClass := get . "storage-class"}}*/}}
{{/*{{- $aclAccountsMap := get . "acl-account-map"}}*/}}

{{/*{{- $freeze := get . "freeze" | default false}}*/}}
{{- $aclConfigmapName := get . "acl-configmap-name"  -}}
{{- $rootPassword := get . "root-password"  -}}
{{- $priorityClassName := get . "priority-class-name"  | default "stateful" -}}

apiVersion: crds.kloudlite.io/v1
kind: HelmChart
metadata:
  name: {{$name}}
  namespace: {{$namespace}}
  labels: {{ $labels | toYAML | nindent 4 }}
  ownerReferences: {{ $ownerRefs | toYAML | nindent 4 }}
spec:
  chartRepo:
    url: https://charts.bitnami.com/bitnami
    name: bitnami
  chartVersion: 18.0.2
  chartName: bitnami/redis

  valuesYaml: |+
    # source: https://github.com/bitnami/charts/tree/main/bitnami/redis/
    global:
     storageClass: {{$storageClass}}
    fullnameOverride: {{$name}}
    image:
      tag: 7.2.1-debian-11-r0

    {{- if $labels }}
    commonLabels: {{ $labels | toYAML | nindent 6 }}
    {{- end}}

    {{- if $nodeSelector }}
    nodeSelector: {{$nodeSelector | toYAML | nindent 6}}
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
      count: 1
      resources:
        requests:
          cpu: {{$requestsCpu}}
          memory: {{$requestsMem}}
        limits:
          cpu: {{ $limitsCpu }}
          memory: {{$limitsMem}}
      persistence:
        enabled: true
        size: {{$storageSize}}

      podLabels: {}
{{/*        {{ if $labels}}{{$labels | toYAML | nindent 6 }}{{ end}}*/}}
{{/*        {{- if .Spec.Region }}*/}}
{{/*        kloudlite.io/region: {{.Spec.Region}}*/}}
{{/*        {{- end }}*/}}
{{/*        kloudlite.io/stateful-node: "true"*/}}

{{/*      priorityClassName: {{$priorityClassName}}*/}}
{{/*      affinity: {{include "NodeAffinity" (dict) | toYAML | nindent 6}}*/}}
{{/*      {{- if .Spec.Region }}*/}}
{{/*      tolerations: {{include "RegionToleration" (dict "region" .Spec.Region) | nindent 6}}*/}}
{{/*      nodeSelector: {{include "RegionNodeSelector" (dict "region" .Spec.Region) | nindent 6}}*/}}
{{/*      {{- end }}*/}}

{{/*    replica:*/}}
{{/*      {{- if $labels}}*/}}
{{/*      podLabels: {{$labels | toYAML | nindent 6 }}*/}}
{{/*      {{- end}}*/}}

{{/*      priorityClassName: {{$priorityClassName}}*/}}
{{/*      affinity: {{include "NodeAffinity" (dict) | toYAML | nindent 6}}*/}}
{{/*      {{- if .Spec.Region}}*/}}
{{/*      tolerations: {{include "RegionToleration" (dict "region" .Spec.Region) | nindent 6}}*/}}
{{/*      nodeSelector: {{include "RegionNodeSelector" (dict "region" .Spec.Region) | nindent 6}}*/}}
{{/*      {{- end }}*/}}

    volumePermissions:
      enabled: true

    metrics:
      enabled: true
