{{ $username := "admin" }} 
{{ $password := "sample" }} 

{{ $redpandaSvcName := .Values.redpandaCluster.name}} 
---
apiVersion: v1
kind: Secret
metadata:
  name: {{.Values.secrets.names.redpandaAdminAuthSecret}}
  namespace: {{.Release.Namespace}}
stringData:
  {{/* USERNAME: {{.Values.redpanda.saslUsername}} */}}
  {{/* PASSWORD: {{.Values.redpanda.saslPassword}} */}}
  {{/* ADMIN_ENDPOINT: {{.Values.redpanda.adminEndpoint}} */}}
  {{/* KAFKA_BROKERS: {{.Values.redpanda.kafkaBrokers}} */}}
  {{/* RPK_ADMIN_FLAGS: {{.Values.redpanda.rpkAdminFlags}} */}}
  {{/* RPK_SASL_FLAGS: {{.Values.redpanda.rpkSaslFlags}} */}}
  USERNAME: {{$username}}
  PASSWORD: {{$password}}
  ADMIN_ENDPOINT: "{{$redpandaSvcName}}.{{.Release.Namespace}}.svc.cluster.local:9644"
  KAFKA_BROKERS: "{{$redpandaSvcName}}.{{.Release.Namespace}}.svc.cluster.local:9092"
  RPK_ADMIN_FLAGS: --user {{$username}} --password {{$password}} --api-urls {{$redpandaSvcName}}.{{.Release.Namespace}}.svc.cluster.local:9644
  RPK_SASL_FLAGS: --user {{$username}} --password {{$password}} --brokers {{$redpandaSvcName}}.{{.Release.Namespace}}.svc.cluster.local:9092 --sasl-mechanism SCRAM-SHA-256 
