apiVersion: redpanda.msvc.kloudlite.io/v1
kind: Admin
metadata:
  name: admin
  namespace: {{.Release.Namespace}}
spec:
  adminEndpoint: {{.Values.redpandaCluster.name}}.{{.Release.Namespace}}.svc.cluster.local:9644
  kafkaBrokers: {{.Values.redpandaCluster.name}}.{{.Release.Namespace}}.svc.cluster.local:9092
