apiVersion: redpanda.msvc.kloudlite.io/v1
kind: Admin
metadata:
  name: admin
  namespace: {{.Release.Namespace}}
spec:
  adminEndpoint: {{.Values.redpandaCluster.name}}.{{.Release.Namespace}}.svc.cluster.local:9644
  {{- $kafkaBrokers := list }} 
  {{- range $i, $e := until (int .Values.redpandaCluster.replicas) }}
  {{- $kafkaBrokers = append $kafkaBrokers (printf "%s-%d.%s.%s.%s:9092" $.Values.redpandaCluster.name $i $.Values.redpandaCluster.name $.Release.Namespace $.Values.clusterInternalDNS) }}
  {{- end }}
  {{- /* kafkaBrokers: {{.Values.redpandaCluster.name}}-0.{{.Values.redpandaCluster.name}}.{{.Release.Namespace}}.svc.cluster.local:9092 */}}
  kafkaBrokers: {{ join "," $kafkaBrokers }}
