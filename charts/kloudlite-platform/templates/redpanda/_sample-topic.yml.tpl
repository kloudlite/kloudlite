apiVersion: cluster.redpanda.com/v1alpha1
kind: Topic
metadata:
  name: sample-topic
  namespace: kl-core
spec:
  partitions: 3
  kafkaApiSpec:
    brokers:
      - redpanda-0.redpanda.kl-core.svc.cluster.local:9092
