apiVersion: redpanda.msvc.kloudlite.io/v1
kind: Topic
metadata:
  name: {{.Values.kafka.topicHarborWebhooks}}
  namespace: {{.Release.Namespace}}
spec:
  redpandaAdmin: admin
  partitionCount: 1
