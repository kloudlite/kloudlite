
apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: {{.Values.apps.webhooksApi.name}}
  namespace: {{.Release.Namespace}}
  annotations:
    kloudlite.io/account-ref: {{.Values.accountName}}
spec:
  accountName: {{.Values.accountName}}
  region: {{.Values.region}}
  {{- if .Values.nodeSelector}}
  nodeSelector: {{.Values.nodeSelector | toYaml | nindent 4}}
  {{- end }}

  {{- if .Values.tolerations }}
  tolerations: {{.Values.tolerations | toYaml | nindent 4}}
  {{- end }}
  
  services:
    - port: 80
      targetPort: 3000
      type: tcp
  containers:
    - name: main
      image: {{.Values.apps.webhooksApi.image}}
      imagePullPolicy: {{.Values.apps.webhooksApi.ImagePullPolicy | default .Values.imagePullPolicy }}
      env:
        - key: CI_ADDR
          value: "{{.Values.apps.ciApi.name}}.{{.Release.Namespace}}.svc.cluster.local:80"

        - key: GITHUB_AUTHZ_SECRET
          type: secret
          refName: "webhook-authz"
          refKey: 'GITHUB_SECRET'

        - key: GITLAB_AUTHZ_SECRET
          type: secret
          refName: "webhook-authz"
          refKey: GITLAB_SECRET

        - key: HARBOR_AUTHZ_SECRET
          type: secret
          refName: "webhook-authz"
          refKey: HARBOR_SECRET

        - key: KL_HOOK_TRIGGER_AUTHZ_SECRET
          type: secret
          refName: "webhook-authz"
          refKey: KLOUDLITE_SECRET

        - key: HTTP_PORT
          value: "3000"

        - key: KAFKA_BROKERS
          type: secret
          refName: {{.Values.redpandaAdminSecretName}}
          refKey: KAFKA_BROKERS

        - key: GIT_WEBHOOKS_TOPIC
          value: {{.Values.redpandaAdminSecretName }}

        - key: HARBOR_WEBHOOK_TOPIC
          value: {{.Values.kafka.topicHarborWebhooks}}

        - key: KAFKA_USERNAME
          type: secret
          refName: {{.Values.redpandaAdminSecretName}}
          refKey: USERNAME

        - key: KAFKA_PASSWORD
          type: secret
          refName: {{.Values.redpandaAdminSecretName}}
          refKey: PASSWORD

      resourceCpu:
        min: "40m"
        max: "60m"
      resourceMemory:
        min: "40Mi"
        max: "60Mi"
---
