{{- if .Values.redpandaCluster.create }}

apiVersion: redpanda.vectorized.io/v1alpha1
kind: Cluster
metadata:
  name: {{.Values.redpandaCluster.name}}
  namespace: {{.Release.Namespace}}
spec:
  image: "vectorized/redpanda"
  version: {{.Values.redpandaCluster.version}}
  replicas: {{.Values.redpandaCluster.replicas}}
  resources: {{.Values.redpandaCluster.resources | toYaml | nindent 4 }}

  {{/* enableSasl: true */}}
  {{/* superusers: */}}
  {{/*   - username: admin */}}

  storage:
    capacity: {{.Values.redpandaCluster.storage.capacity}}
    {{/* Note: it is recommended to use XFS */}}
    {{- if .Values.persistence.storageClasses.xfs}}
    storageClassName: {{.Values.persistence.storageClasses.xfs}}
    {{- else if .Values.persistence.storageClasses.ext4 }}
    storageClassName: {{.Values.persistence.storageClasses.ext4}}
    {{- end }}

  configuration:
    rpcServer:
      port: 33145
    kafkaApi:
    - port: 9092

    pandaproxyApi:
    - port: 8082
    schemaRegistry:
      port: 8081
    adminApi:
    - port: 9644
    developerMode: true

{{- end }}
