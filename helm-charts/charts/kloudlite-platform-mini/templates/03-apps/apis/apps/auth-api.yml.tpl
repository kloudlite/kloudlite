{{- $appName := include "apps.authApi.name" . }}
apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: {{ $appName }}
  namespace: {{.Release.Namespace}}
  labels: {{ include "common.pod-labels" . | nindent 4 }}
  annotations:
    kloudlite.io/checksum.oauth2-secrets: {{ include (print $.Template.BasePath "/03-apps/apis/secrets/oauth-secrets.yml.tpl") . | sha256sum }}
    {{ include "common.pod-annotations" . | nindent 4 }}
spec:
  serviceAccount: {{ .Values.serviceAccounts.normal.name }}

  nodeSelector: {{ .Values.scheduling.stateless.nodeSelector | toYaml | nindent 4 }}
  tolerations: {{ .Values.scheduling.stateless.tolerations | toYaml | nindent 4 }}
  
  topologySpreadConstraints:
    {{- include "tsc-hostname" (dict "kloudlite.io/app.name" $appName) | nindent 4 }}

  replicas: {{.Values.apps.authApi.replicas }}
  services:
    - port: 3000
    - port: 3001

  hpa:
    enabled: {{.Values.apps.authApi.hpa.enabled}}
    minReplicas: {{.Values.apps.authApi.hpa.minReplicas}}
    maxReplicas: {{.Values.apps.authApi.hpa.maxReplicas}}
    thresholdCpu: 70
    thresholdMemory: 80

  containers:
    - name: main
      image: '{{.Values.apps.authApi.image.repository}}:{{.Values.apps.authApi.image.tag | default (include "image-tag" .) }}'
      imagePullPolicy: {{ include "image-pull-policy" .}}
      resourceCpu:
        min: "50m"
        max: "60m"
      resourceMemory:
        min: "80Mi"
        max: "120Mi"
      env:
        - key: PORT
          value: "3000"

        - key: GRPC_PORT
          value: "3001"

        - key: MONGO_URI
          type: secret
          refName: {{.Values.mongo.secretKeyRef.name}}
          refKey: {{.Values.mongo.secretKeyRef.key}}

        - key: MONGO_DB_NAME
          value: "auth-db"

        - key: COMMS_SERVICE
          value: "comms:3001"

        - key: SESSION_KV_BUCKET
          value: {{.Values.nats.buckets.sessionKV}}

        - key: RESET_PASSWORD_TOKEN_KV_BUCKET
          value: {{.Values.nats.buckets.resetTokenKV}}

        - key: VERIFY_TOKEN_KV_BUCKET
          value: {{.Values.nats.buckets.verifyTokenKV}}

        - key: NATS_URL
          value: {{.Values.nats.url}}

        - key: ORIGINS
          {{- /* FIXME: this should be from baseDomains */}}
          value: "https://kloudlite.io,https://studio.apollographql.com"

        - key: COOKIE_DOMAIN
          value: "{{- include "kloudlite.cookie-domain" . }}"

        - key: USER_EMAIL_VERIFICATION_ENABLED
          value: {{ (or (not .Values.sendgrid.apiKey) (not .Values.sendgrid.sender)) | ternary false true | squote }}

        {{- if .Values.apps.authApi.oAuth2.providers.github.enabled }}
        - key: GITHUB_APP_PK_FILE
          value: /github/github-app-pk.pem
        {{- end }}

      envFrom:
        - type: secret
          refName: {{ include "apps.authApi.oAuth2-secret.name" . }}

      {{- if .Values.apps.authApi.oAuth2.providers.github.enabled }}
      volumes:
        - mountPath: /github
          type: secret
          refName: {{ include "apps.authApi.oAuth2-secret.name" . }}
          items:
            - key: github-app-pk.pem
              fileName: github-app-pk.pem
      {{- end }}

      livenessProbe: &probe
        type: httpGet
        httpGet:
          path: /_healthy
          port: {{ include "apps.authApi.httpPort" . }}
        initialDelay: 3
        interval: 10

      readinessProbe: *probe
