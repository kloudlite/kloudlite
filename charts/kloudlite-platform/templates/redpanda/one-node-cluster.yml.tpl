apiVersion: redpanda.vectorized.io/v1alpha1
kind: Cluster
metadata:
  name: {{.Values.redpandaCluster.name}}
  namespace: {{.Release.Namespace}}
spec:
  image: "vectorized/redpanda"
  version: {{.Values.redpandaCluster.version}}
  replicas: {{.Values.redpandaCluster.replicas}}
  resources: {{.Values.redpandaCluster.resources | toPrettyJson }}

  {{/* enableSasl: true */}}
  {{/* superusers: */}}
  {{/*   - username: admin */}}

  storage:
    capacity: {{.Values.redpandaCluster.storage.capacity}}
    {{/* Note: XFS */}}
    {{ if .Values.redpandaCluster.storage.storageClassName -}}
    storageClassName: {{.Values.redpandaCluster.storage.storageClassName}}
    {{ end }}

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

