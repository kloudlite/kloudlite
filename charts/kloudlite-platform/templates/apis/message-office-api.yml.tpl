apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: {{.Values.apps.messageOfficeApi.name}}
  namespace: {{.Release.Namespace}}
  annotations:
    kloudlite.io/account-ref: {{.Values.accountName}}
spec:
  region: {{.Values.region | default ""}}
  {{/* serviceAccount: {{.Values.normalSvcAccount}} */}}
  serviceAccount: {{.Values.clusterSvcAccount}}

  {{ include "node-selector-and-tolerations" . | nindent 2 }}

  services:
    - port: 80
      targetPort: 3000
      name: http
      type: tcp

    - port: 3001
      targetPort: 3001
      name: grpc
      type: tcp

  containers:
    - name: main
      image: {{.Values.apps.messageOfficeApi.image}}
      imagePullPolicy: {{.Values.apps.messageOfficeApi.ImagePullPolicy | default .Values.imagePullPolicy }}
      resourceCpu:
        min: "50m"
        max: "100m"
      resourceMemory:
        min: "50Mi"
        max: "100Mi"
      env:
        - key: HTTP_PORT
          value: "3000"

        - key: GRPC_PORT
          value: '3001'
      
        - key: DB_URI
          type: secret
          refName: "mres-{{.Values.managedResources.messageOfficeDb}}"
          refKey: URI

        - key: DB_NAME
          value: {{.Values.managedResources.messageOfficeDb}}

        - key: AUTH_REDIS_HOSTS
          type: secret
          refName: "mres-{{.Values.managedResources.authRedis}}"
          refKey: HOSTS

        - key: AUTH_REDIS_PASSWORD
          type: secret
          refName: "mres-{{.Values.managedResources.authRedis}}"
          refKey: PASSWORD

        - key: AUTH_REDIS_PREFIX
          type: secret
          refName: "mres-{{.Values.managedResources.authRedis}}"
          refKey: PREFIX

        - key: AUTH_REDIS_USERNAME
          type: secret
          refName: "mres-{{.Values.managedResources.authRedis}}"
          refKey: USERNAME

        - key: KAFKA_TOPIC_STATUS_UPDATES
          value: {{.Values.kafka.topicStatusUpdates}}

        - key: KAFKA_TOPIC_INFRA_UPDATES
          value: {{.Values.kafka.topicInfraStatusUpdates}}

        - key: KAFKA_TOPIC_ERROR_ON_APPLY
          value: {{.Values.kafka.topicErrorOnApply}}

        - key: KAFKA_TOPIC_BYOC_CLIENT_UPDATES
          value: {{.Values.kafka.topicBYOCClientUpdates}}

        - key: KAFKA_BROKERS
          type: secret
          refName: "{{.Values.secrets.names.redpandaAdminAuthSecret}}"
          refKey: KAFKA_BROKERS

        - key: KAFKA_SASL_USERNAME
          type: secret
          refName: "{{.Values.secrets.names.redpandaAdminAuthSecret}}"
          refKey: USERNAME

        - key: KAFKA_SASL_PASSWORD
          type: secret
          refName: "{{.Values.secrets.names.redpandaAdminAuthSecret}}"
          refKey: PASSWORD

        - key: KAFKA_CONSUMER_GROUP
          value: {{.Values.kafka.consumerGroupId}}
