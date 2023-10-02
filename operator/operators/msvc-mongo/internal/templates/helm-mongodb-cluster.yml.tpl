{{- $name := get . "name" }} 
{{- $namespace := get . "namespace" }} 
{{- $labels := get . "labels" | default dict }} 
{{- $ownerRefs := get . "owner-refs" | default list }}

{{- $storageClass := get . "storage-class" }}
{{- $storageSize := get . "storage-size" }}

{{- $replicaCount := get . "replica-count" }}

---

apiVersion: crds.kloudlite.io/v1
kind: HelmChart
metadata:
  name: {{$name}}
  namespace: {{$namespace}}
spec:
  chartRepo:
    name: bitnami
    url: https://charts.bitnami.com/bitnami
  chartName: bitnami/mongodb
  chartVersion: 14.0.2
  valuesYaml: |+
    global:
      storageClass: {{$storageClass}}
    image:
      registry: docker.io
      repository: bitnami/mongodb
      tag: 7.0.2-debian-11-r0

    fullnameOverride: {{.Name}}

    architecture: "replicaset"

    replicaCount: {{ index (mustFromJson (.Spec.Inputs | toJson)) "replica_count" | default 1 | int64 }}

    auth:
      enabled: true
      rootUser: {{$rootUser}}
      rootPassword: {{ index (mustFromJson (.Spec.Inputs | toJson)) "root_password" | quote }}
      replicaSetKey: {{ index (mustFromJson (.Spec.Inputs | toJson)) "replica_set_key" | quote }}

    replicaSetName: rs
    replicaSetHostnames: true

    directoryPerDB: true

    persistence:
      enabled: true
      size: {{$storageSize}}

    volumePermissions:
      enabled: true

---

apiVersion: msvc.kloudlite.io/v1
kind: HelmMongoDB
metadata:
  name: {{$name}}
  namespace: {{$namespace}}
  labels: {{$labels | toYaml | nindent 4}}
  ownerReferences: {{$ownerRefs | toYaml | nindent 4}}
spec:
  global:
    storageClass: {{$storageClass}}
  image:
    repository: bitnami/mongodb
    tag: 5.0.8-debian-10-r20

  fullnameOverride: {{.Name}}

  architecture: "replicaset"
  replicaCount: {{ index (mustFromJson (.Spec.Inputs | toJson)) "replica_count" | default 1 | int64 }}
  replicaSetName: rs
  replicaSetHostnames: true

  auth:
    enabled: true
    rootPassword: {{ index (mustFromJson (.Spec.Inputs | toJson)) "root_password" | quote }}
    replicaSetKey: {{ index (mustFromJson (.Spec.Inputs | toJson)) "replica_set_key" | quote }}

  persistence:
    enabled: true
    size: {{$storageSize}}

  volumePermissions:
    enabled: true

#  metrics:
#    enabled: true

  resources:
    requests:
      cpu: {{ index (mustFromJson (.Spec.Inputs | toJson)) "cpu_min" | default 400 }}m
      memory: {{ index (mustFromJson (.Spec.Inputs | toJson)) "memory_min" | default 400 }}Mi
    limits:
      cpu: {{ index (mustFromJson (.Spec.Inputs | toJson)) "cpu_max" |  default 400}}m
      memory: {{ index (mustFromJson (.Spec.Inputs | toJson)) "memory_max" | default 500}}Mi

