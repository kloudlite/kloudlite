apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: {{.Values.apps.auditLoggingWorker.name}}
  namespace: {{.Release.Namespace}}
spec:
{{/*  region: {{.Values.region | default ""}}*/}}
  serviceAccount: {{.Values.normalSvcAccount}}

  {{ include "node-selector-and-tolerations" . | nindent 2 }}

  services: []
  containers:
    - name: main
      image: {{.Values.apps.auditLoggingWorker.image}}
      imagePullPolicy: {{.Values.apps.auditLoggingWorker.imagePullPolicy | default .Values.imagePullPolicy }}
      resourceCpu:
        min: "50m"
        max: "70m"
      resourceMemory:
        min: "50Mi"
        max: "70Mi"
      env:
        - key: KAFKA_BROKERS
          type: secret
{{/*          refName: {{.Values.redpandaAdminSecretName}}*/}}
          refName: {{.Values.secretNames.redpandaAdminAuthSecret}}
          refKey: "KAFKA_BROKERS"

        - key: KAFKA_USERNAME
          type: secret
{{/*          refName: {{.Values.redpandaAdminSecretName}}*/}}
          refName: {{.Values.secretNames.redpandaAdminAuthSecret}}
          refKey: "USERNAME"

        - key: KAFKA_PASSWORD
          type: secret
{{/*          refName: {{.Values.redpandaAdminSecretName}}*/}}
          refName: {{.Values.secretNames.redpandaAdminAuthSecret}}
          refKey: "PASSWORD"

        - key: KAFKA_SUBSCRIPTION_TOPICS
          value: {{.Values.kafka.auditEvents}}

        - key: KAFKA_CONSUMER_GROUP_ID
          value: {{.Values.kafka.consumerGroupId}}

        - key: EVENTS_DB_URI
          type: secret
          refName: {{printf "mres-%s" .Values.managedResources.eventsDb}}
          refKey: URI

        - key: EVENTS_DB_NAME
          value: {{.Values.managedResources.eventsDb}}
