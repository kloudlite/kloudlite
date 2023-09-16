{{- $name := get . "name" }}
{{- $namespace := get . "namespace" }}
{{- $labels := get . "labels" | default dict }}
{{- $ownerRefs := get . "owner-refs" | default list }}

{{- $storageSize := get . "storage-size" }}
{{- $storageClass := get . "storage-class" }}

{{- $requestsCpu := get . "requests-cpu" }}
{{- $requestsMem := get . "requests-mem" }}
{{- $limitsCpu := get . "limits-cpu" }}
{{- $limitsMem := get . "limits-mem" }}

{{- $freeze := get . "freeze" | default false}}
{{/*{{- $priorityClassName := get . "priority-class-name"  | default "stateful" -}}*/}}
{{- $existingSecret := get . "existing-secret" -}}


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
  chartVersion: 13.18.1
  chartName: bitnami/mongodb

  valuesYaml: |+
    # source: https://github.com/bitnami/charts/tree/main/bitnami/mongodb/
    global:
      storageClass: {{$storageClass}}
    fullnameOverride: {{$name}}
    image:
      tag: 5.0.8-debian-10-r20

    architecture: standalone
    useStatefulSet: true

    replicaCount: {{ if $freeze}}0{{else}}1{{end}}

    commonLabels: {{$labels | toYAML | nindent 6}}
    podLabels: {{$labels | toYAML | nindent 6}}

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
