{{- $name := get . "name" }} 
{{- $namespace := get . "namespace" }} 

{{- $releaseName := get . "release-name" }}

{{- $labels := get . "labels" | default dict }} 
{{- $annotations := get . "annotations" | default dict}}
{{- $nodeSelector := get . "node-selector" |default dict }}
{{- $tolerations := get . "tolerations" | default list }}
{{- $priorityClassname := get . "priority-classname" | default "stateful" }}

{{- $podLabels := get . "pod-labels" }}
{{- $podAnnotations := get . "pod-annotations" }}

{{- $topologySpreadConstraints := get . "topology-spread-constraints" }}

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
  chartRepoURL: https://charts.bitnami.com/bitnami
  chartName: mongodb
  chartVersion: 14.3.0

  {{- if $releaseName }}
  releaseName: {{$releaseName}}
  {{- end }}

  jobVars:
    tolerations: {{$tolerations | toYAML | nindent 6 }}
    nodeSelector: {{$nodeSelector | toYAML | nindent 6 }}

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

    commonLabels: {{$labels | toYAML | nindent 6}}

    podLabels: {{$podLabels | toYAML | nindent 6}}
    podAnnotations: {{$podAnnotations | toYAML | nindent 6}}

    directoryPerDB: true

    tolerations: {{$tolerations | toYAML | nindent 8 }}
    nodeSelector: {{$nodeSelector | toYAML | nindent 8 }}

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

    priorityClassName: "{{$priorityClassname}}"

    topologySpreadConstraints:
      - maxSkew: 1
        topologyKey: kloudlite.io/provider.az
        whenUnsatisfiable: DoNotSchedule
        nodeAffinityPolicy: Honor
        nodeTaintsPolicy: Honor
        labelSelector:
          matchLabels: {{$labels | toYAML | nindent 12}}
      - maxSkew: 1
        topologyKey: kloudlite.io/node.name
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
