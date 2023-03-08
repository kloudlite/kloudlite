---
apiVersion: v1
kind: Secret
metadata:
  name: {{.Values.secrets.names.redpandaAdminAuthSecret}}
  namespace: {{.Release.Namespace}}
stringData:
  USERNAME: {{.Values.redpanda.saslUsername}}
  PASSWORD: {{.Values.redpanda.saslPassword}}
  ADMIN_ENDPOINT: {{.Values.redpanda.adminEndpoint}}
  KAFKA_BROKERS: {{.Values.redpanda.kafkaBrokers}}
  RPK_ADMIN_FLAGS: {{.Values.redpanda.rpkAdminFlags}}
  RPK_SASL_FLAGS: {{.Values.redpanda.rpkSaslFlags}}
