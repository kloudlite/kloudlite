{{- /* NOTE: sourced from: https://github.com/strimzi/strimzi-kafka-operator/blob/main/examples/kafka/nodepools/kafka-with-dual-role-kraft-nodes.yaml */}}
{{- if .Values.managedServices.kafkaSvc.enabled }}

{{- $kafkaClusterName := "kafka-cluster-name" }} 

apiVersion: kafka.strimzi.io/v1beta2
kind: KafkaNodePool
metadata:
  name: {{$kafkaClusterName}}-dual-role
  labels:
    strimzi.io/cluster: {{$kafkaClusterName}}
spec:
  replicas: 3
  roles:
    - controller
    - broker
  storage:
    type: persistent-claim
    class: sc-ext4
    size: 5Gi
    deleteClaim: false
---

apiVersion: kafka.strimzi.io/v1beta2
kind: Kafka
metadata:
  name: {{$kafkaClusterName}}
  annotations:
    strimzi.io/node-pools: enabled
spec:
  kafka:
    version: 3.5.1
    # The replicas field is required by the Kafka CRD schema while the KafkaNodePools feature gate is in alpha phase.
    # But it will be ignored when Kafka Node Pools are used
    replicas: 3
    listeners:
      - name: plain
        port: 9092
        type: internal
        tls: false
      - name: tls
        port: 9093
        type: internal
        tls: true
    config:
      offsets.topic.replication.factor: 3
      transaction.state.log.replication.factor: 3
      transaction.state.log.min.isr: 2
      default.replication.factor: 3
      min.insync.replicas: 2
      inter.broker.protocol.version: "3.5"
    # The storage field is required by the Kafka CRD schema while the KafkaNodePools feature gate is in alpha phase.
    # But it will be ignored when Kafka Node Pools are used
    storage:
      type: persistent-claim
      class: sc-ext4
  # The ZooKeeper section is required by the Kafka CRD schema while the UseKRaft feature gate is in alpha phase.
  # But it will be ignored when running in KRaft mode
  zookeeper:
    replicas: 3
    storage:
      type: persistent-claim
      size: 100Gi
      deleteClaim: false
  {{- /* entityOperator: */}}
  {{- /*   userOperator: {} */}}

{{- end }}
