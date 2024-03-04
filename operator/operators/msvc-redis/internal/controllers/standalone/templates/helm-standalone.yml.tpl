{{- $name := get . "name" }}
{{- $namespace := get . "namespace" }}

{{- $labels := get . "labels" | default dict }}
{{- $ownerRefs := get . "owner-refs" | default list }}

{{- $nodeSelector := get . "node-selector" | default dict }}
{{- $tolerations := get . "tolerations" | default list }}

{{- $storageSize := get . "storage-size" }}
{{- $storageClass := get . "storage-class" }}

{{- $requestsCpu := get . "requests-cpu" }}
{{- $requestsMem := get . "requests-mem" }}

{{- $limitsCpu := get . "limits-cpu" }}
{{- $limitsMem := get . "limits-mem" }}

{{/*{{- $freeze := get . "freeze" | default false}}*/}}
{{- /* {{- $aclConfigmapName := get . "acl-configmap-name"  -}} */}}
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
  chartRepoURL: https://charts.bitnami.com/bitnami
  chartVersion: 18.0.2
  chartName: redis

  values:
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

    {{- /* existingConfigmap: {{$aclConfigmapName}} */}}

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

      podLabels: {{$labels | toYAML | nindent 8 }}

    priorityClassName: {{$priorityClassName}}

    volumePermissions:
      enabled: true

    metrics:
      enabled: true
