apiVersion: redpanda.msvc.kloudlite.io/v1
kind: Topic
metadata:
  name: {{.Values.kafka.topicInfraStatusUpdates}}
  namespace: {{.Release.Namespace}}
spec:
  redpandaAdmin: admin
  partitionCount: 1
