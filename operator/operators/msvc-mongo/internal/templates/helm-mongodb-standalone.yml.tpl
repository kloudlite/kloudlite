{{- $name := get . "name" }}
{{- $namespace := get . "namespace" }}
{{- $labels := get . "labels" | default dict }}
{{- $annotations := get . "annotations" | default dict}}
{{- $ownerRefs := get . "owner-refs" | default list }}

{{- $nodeSelector := get . "node-selector" | default dict }}

{{- $storageSize := get . "storage-size" }}
{{- $storageClass := get . "storage-class" }}

{{- $requestsCpu := get . "requests-cpu" }}
{{- $requestsMem := get . "requests-mem" }}
{{- $limitsCpu := get . "limits-cpu" }}
{{- $limitsMem := get . "limits-mem" }}

{{- $priorityClassName := get . "priority-class-name"  | default "stateful" -}}
{{- $existingSecret := get . "existing-secret" -}}


apiVersion: crds.kloudlite.io/v1
kind: HelmChart
metadata:
  name: {{$name}}
  namespace: {{$namespace}}
  labels: {{ $labels | toYAML | nindent 4 }}
  annotations: {{$annotations | toYAML | nindent 4}}
  ownerReferences: {{ $ownerRefs | toYAML | nindent 4 }}
spec:
  chartRepoURL: https://charts.bitnami.com/bitnami
  chartVersion: 13.18.1
  chartName: mongodb

  values:
    # source: https://github.com/bitnami/charts/tree/main/bitnami/mongodb/
    global:
      {{- if $storageClass }}
      storageClass: {{$storageClass}}
      {{- end }}
    fullnameOverride: {{$name}}
    image:
      tag: 5.0.8-debian-10-r20

    architecture: standalone
    useStatefulSet: true

    replicaCount: 1

    commonLabels: {{$labels | toYAML | nindent 6}}
    podLabels: {{$labels | toYAML | nindent 6}}
    nodeSelector: {{$nodeSelector | toYAML | nindent 6 }}

    auth:
      enabled: true
      existingSecret: {{$existingSecret}}

    persistence:
      enabled: true
      size: {{$storageSize}}

    volumePermissions:
      enabled: true

    metrics:
      enabled: true

    priorityClassName: "{{$priorityClassName}}"

    livenessProbe:
      enabled: true
      timeoutSeconds: 15

    readinessProbe:
      enabled: true
      initialDelaySeconds: 10
      periodSeconds: 30
      timeoutSeconds: 20

    resources:
      requests:
        cpu: {{$requestsCpu}}
        memory: {{$requestsMem}}
      limits:
        cpu: {{$requestsCpu}}
        memory: {{$requestsMem}}
