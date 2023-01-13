{{- $namespace := get . "namespace" -}}
{{- $kafkaBrokers := get . "kafka-brokers" -}}
{{- $kafkaSaslUsername := get . "kafka-sasl-username" -}}
{{- $kafkaSaslPassword := get . "kafka-sasl-password" -}}
{{- $clusterId := get . "cluster-id" -}}

{{- $kafkaTopicStatusUpdates := get . "kafka-topic-status-updates" -}}
{{- $kafkaTopicBillingUpdates := get . "kafka-topic-billing-updates" -}}
{{- $kakfaTopicPipelineRunUpdates := get . "kakfa-topic-pipeline-run-updates" -}}

apiVersion: v1
kind: Secret
metadata:
  name: "status-n-billing-env"
  namespace: {{$namespace}}
stringData:
  KAFKA_BROKERS: {{$kafkaBrokers}}
  KAFKA_SASL_USERNAME: {{$kafkaSaslUsername}}
  KAFKA_SASL_PASSWORD: {{$kafkaSaslPassword}}
  CLUSTER_ID: {{$clusterId}}

  KAFKA_TOPIC_STATUS_UPDATES: {{$kafkaTopicStatusUpdates}}
  KAFKA_TOPIC_BILLING_UPDATES: {{$kafkaTopicBillingUpdates}}
  KAFKA_TOPIC_PIPELINE_RUN_UPDATES: {{$kakfaTopicPipelineRunUpdates}}
