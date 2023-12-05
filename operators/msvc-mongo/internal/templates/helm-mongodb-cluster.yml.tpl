{{- $name := get . "name" }} 
{{- $namespace := get . "namespace" }} 
{{- $labels := get . "labels" | default dict }} 
{{- $annotations := get . "annotations" | default dict}}
{{- $ownerRefs := get . "owner-refs" | default list }}

{{- $storageClass := get . "storage-class" }}
{{- $storageSize := get . "storage-size" }}

{{- $replicaCount := get . "replica-count" }}
{{- $rootUser := get . "root-user" }}
{{- $authExistingSecret := get . "auth-existing-secret" }}

{{- $cpuMin := get . "cpu-min" }}
{{- $cpuMax := get . "cpu-max" }}
{{- $memoryMin := get . "memory-min" }}
{{- $memoryMax := get . "memory-max" }}

---

apiVersion: crds.kloudlite.io/v1
kind: HelmChart
metadata:
  name: {{$name}}
  namespace: {{$namespace}}
  labels: {{$labels | toYAML | nindent 4}}
  annotations: {{$annotations | toYAML | nindent 4}}
  ownerReferences: {{$ownerRefs | toYAML | nindent 4}}
spec:
  chartRepo:
    name: bitnami
    url: https://charts.bitnami.com/bitnami
  chartName: bitnami/mongodb
  chartVersion: 14.3.1

  values:
    # source: https://github.com/bitnami/charts/tree/main/bitnami/mongodb/
    global:
      storageClass: {{$storageClass}}

    architecture: "replicaset"
    image:
      registry: docker.io
      repository: bitnami/mongodb
      tag: 7.0.2-debian-11-r0

    fullnameOverride: {{$name}}

    replicaCount: {{ $replicaCount | int64 }}
    replicaSetName: rs
    replicaSetHostnames: true
    podLabels: {{$labels | toYAML | nindent 6}}

    directoryPerDB: true

    persistence:
      enabled: true
      size: {{$storageSize}}

    auth:
      enabled: true
      rootUser: {{$rootUser}}
      existingSecret: {{$authExistingSecret}}

    volumePermissions:
      enabled: true

    metrics:
      enabled: true

    priorityClassName: "stateful"

    topologySpreadConstraints:
      - maxSkew: 1
        topologyKey: kloudlite.io/provider.az
        whenUnsatisfiable: DoNotSchedule
        nodeAffinityPolicy: Honor
        nodeTaintsPolicy: Honor
        labelSelector:
          matchLabels: {{$labels | toYAML | nindent 12}}
    
    resources:
      requests:
        cpu: {{$cpuMin}}
        memory: {{$memoryMin}}
      limits:
        cpu: {{$cpuMax}}
        memory: {{$memoryMax}}
