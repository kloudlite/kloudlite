{{- $app := .Values.apps.messagesDistributionWorker }} 

{{- if $app.enabled }}
apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: {{$app.name}}
  namespace: {{.Release.Namespace}}
spec:
  serviceAccount: {{.Values.normalSvcAccount}}

  {{ include "node-selector-and-tolerations" . | nindent 2 }}

  services: []
  containers:
    - name: main
      image: {{$app.image}}
      imagePullPolicy: {{$app.imagePullPolicy | default .Values.imagePullPolicy }}
      resourceCpu:
        min: "50m"
        max: "70m"
      resourceMemory:
        min: "50Mi"
        max: "70Mi"
      env:
        - key: KAFKA_BROKERS
          type: secret
          refName: {{.Values.secretNames.redpandaAdminAuthSecret}}
          refKey: "KAFKA_BROKERS"

        - key: WAIT_QUEUE_KAFKA_TOPIC
          value: {{.Values.kafka.topicSendMessagesToTargetWaitQueue}}
        
        - key: WAIT_QUEUE_KAFKA_CONSUMER_GROUP
          value: {{$app.configuration.kafkaConsumerGroupId}}

        - key: REDPANDA_HTTP_ADDR
          value: {{.Values.redpandaCluster.name}}.{{.Release.Namespace}}.svc.{{.Values.clusterInternalDNS}}:8082

        - key: NEW_TOPIC_PARTITIONS_COUNT
          value: {{$app.configuration.newTopicPartitionCount | squote}}

        - key: NEW_TOPIC_REPLICATION_COUNT
          value: {{$app.configuration.newTopicReplicationCount | squote}}
{{- end }}
