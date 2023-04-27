
apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: {{.Values.apps.webhooksApi.name}}
  namespace: {{.Release.Namespace}}
  annotations:
    kloudlite.io/account-ref: {{.Values.accountName}}
spec:
  region: {{.Values.region | default ""}}
  serviceAccount: {{.Values.normalSvcAccount}}

  {{ include "node-selector-and-tolerations" . | nindent 2 }}

  services:
    - port: 80
      targetPort: 3000
      type: tcp
  containers:
    - name: main
      image: {{.Values.apps.webhooksApi.image}}
      imagePullPolicy: {{.Values.apps.webhooksApi.ImagePullPolicy | default .Values.imagePullPolicy }}
      env:
        - key: HARBOR_AUTHZ_SECRET
          type: secret
          refName: {{.Values.secrets.names.webhookAuthzSecret}}
          refKey: HARBOR_SECRET

        - key: KL_HOOK_TRIGGER_AUTHZ_SECRET
          type: secret
          refName: {{.Values.secrets.names.webhookAuthzSecret}}
          refKey: KLOUDLITE_SECRET

        - key: HTTP_PORT
          value: "3000"

        - key: KAFKA_BROKERS
          type: secret
          refName: {{.Values.secrets.names.redpandaAdminAuthSecret}}
          refKey: KAFKA_BROKERS

        - key: HARBOR_WEBHOOK_TOPIC
          value: {{.Values.kafka.topicHarborWebhooks}}

        - key: KAFKA_USERNAME
          type: secret
          refName: {{.Values.secrets.names.redpandaAdminAuthSecret}}
          refKey: USERNAME

        - key: KAFKA_PASSWORD
          type: secret
          refName: {{.Values.secrets.names.redpandaAdminAuthSecret}}
          refKey: PASSWORD

      resourceCpu:
        min: "40m"
        max: "60m"
      resourceMemory:
        min: "40Mi"
        max: "60Mi"
---
