{{- define "resource-watcher-env" -}}
- name: ACCOUNT_NAME
  value: {{.Values.accountName}}

- name: CLUSTER_NAME
  value: {{.Values.clusterName}}

- name: KAFKA_BROKERS
  value: redpanda.kl-core.svc.{{.Values.clusterInternalDNS}}:9092

- name: KAFKA_RESOURCE_UPDATES_TOPIC
  value: {{.Values.kafka.topicStatusUpdates}}

- name: KAFKA_INFRA_UPDATES_TOPIC
  value: {{.Values.kafka.topicInfraStatusUpdates}}
{{- end -}}
