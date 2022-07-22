apiVersion: msvc.kloudlite.io/v1
kind: HelmMongoDB
metadata:
  name: {{.Name}}
  namespace: {{.Namespace}}
  ownerReferences:
    - apiVersion: {{.APIVersion}}
      kind: {{.Kind}}
      name: {{.Name}}
      uid: {{.UID}}
      controller: true
      blockOwnerDeletion: true
spec:
#  global:
#    storageClass: do-block-storage
  image:
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
    size: {{ index (mustFromJson (.Spec.Inputs | toJson)) "size" | default "1Gi" | quote }}

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
