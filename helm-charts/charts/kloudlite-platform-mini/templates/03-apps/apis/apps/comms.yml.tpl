{{- $appName := "comms" }}

apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: "{{$appName}}"
  namespace: {{.Release.Namespace}}
  labels: {{ include "common.pod-labels" . | nindent 4 }}
  annotations: {{ include "common.pod-annotations" . | nindent 4 }}
spec:
  serviceAccount: {{ .Values.serviceAccounts.normal.name }}

  nodeSelector: {{.Values.scheduling.stateless.nodeSelector | toYaml | nindent 4}}
  tolerations: {{.Values.scheduling.stateless.tolerations | toYaml | nindent 4}}

  topologySpreadConstraints:
    {{ include "tsc-hostname" (dict "kloudlite.io/app.name" $appName) | nindent 4 }}

  replicas: {{.Values.apps.commsApi.replicas }}

  services:
    - port: {{ include "apps.commsApi.httpPort" . }}
    - port: {{ include "apps.commsApi.grpcPort" . }}

  hpa:
    enabled: {{.Values.apps.commsApi.hpa.enabled}}
    minReplicas: {{.Values.apps.commsApi.hpa.minReplicas}}
    maxReplicas: {{.Values.apps.commsApi.hpa.maxReplicas}}
    thresholdCpu: 70
    thresholdMemory: 80

  containers:
    - name: main
      image: '{{.Values.apps.commsApi.image.repository}}:{{.Values.apps.commsApi.image.tag | default (include "image-tag" .) }}'
      imagePullPolicy: {{ include "image-pull-policy" .}}
      resourceCpu:
        min: "50m"
        max: "80m"
      resourceMemory:
        min: "80Mi"
        max: "120Mi"

      env:
        - key: HTTP_PORT
          value: {{ include "apps.commsApi.httpPort" . | squote }}

        - key: GRPC_PORT
          value: {{ include "apps.commsApi.grpcPort" .  | squote}}

        - key: SUPPORT_EMAIL
          value: {{.Values.sendgrid.sender}}

        - key: SENDGRID_API_KEY
          value: {{.Values.sendgrid.apiKey}}

        - key: ACCOUNTS_WEB_INVITE_URL
          value: https://auth.{{ .Values.webHost }}/invite

        - key: PROJECTS_WEB_INVITE_URL
          value: https://auth.{{ .Values.webHost }}/invite

        - key: KLOUDLITE_CONSOLE_WEB_URL
          value: https://console.{{.Values.webHost}}

        - key: RESET_PASSWORD_WEB_URL
          value: https://auth.{{.Values.webHost}}/reset-password

        - key: VERIFY_EMAIL_WEB_URL
          value: https://auth.{{.Values.webHost }}/verify-email
        
        {{/* TODO: url should definitely NOT be auth.Values.webHost */}}
        - key: EMAIL_LINKS_BASE_URL
          value: https://auth.{{.Values.webHost}}/

        {{- /* notifications params */}}
        - key: ACCOUNT_COOKIE_NAME
          value: {{ include "kloudlite.account-cookie-name" . }}

        - key: NATS_URL
          value: {{.Values.nats.url}}

        - key: NOTIFICATION_NATS_STREAM
          value: {{.Values.nats.streams.events}}

        - key: SESSION_KV_BUCKET
          value: {{.Values.nats.buckets.sessionKV}}
        
        - key: IAM_GRPC_ADDR
          value: 'iam:{{ include "apps.iamApi.grpcPort" . }}'

        - key: MONGO_URI
          type: secret
          refName: {{.Values.mongo.secretKeyRef.name}}
          refKey: {{.Values.mongo.secretKeyRef.key}}

        - key: MONGO_DB_NAME
          value: "comms-db"
