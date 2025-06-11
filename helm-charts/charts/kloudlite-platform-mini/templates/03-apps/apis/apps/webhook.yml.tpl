{{- $appName := include "apps.webhooksApi.name" . }}

apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: {{ $appName }}
  namespace: {{.Release.Namespace}}
  labels: {{ include "common.pod-labels" . | nindent 4}}
  annotations: {{ include "common.pod-annotations" . | nindent 4}}
spec:
  serviceAccount: {{.Values.serviceAccounts.normal.name}}

  nodeSelector: {{.Values.scheduling.stateless.nodeSelector | toYaml | nindent 4}}
  tolerations: {{.Values.scheduling.stateless.tolerations | toYaml | nindent 4}}
  
  topologySpreadConstraints:
    {{ include "tsc-hostname" (dict "kloudlite.io/app.name" $appName) | nindent 4 }}

  replicas: {{.Values.apps.webhooksApi.replicas}}

  services:
    - port: {{ include "apps.webhooksApi.httpPort" . }}

  hpa:
    enabled: {{.Values.apps.webhooksApi.hpa.enabled}}
    minReplicas: {{.Values.apps.webhooksApi.hpa.minReplicas}}
    maxReplicas: {{.Values.apps.webhooksApi.hpa.maxReplicas}}
    thresholdCpu: 70
    thresholdMemory: 80

  containers:
    - name: main
      image: '{{.Values.apps.webhooksApi.image.repository}}:{{.Values.apps.webhooksApi.image.tag | default (include "image-tag" .) }}'
      imagePullPolicy: {{ include "image-pull-policy" .}}
      resourceCpu:
        min: "30m"
        max: "50m"
      resourceMemory:
        min: "50Mi"
        max: "100Mi"
      env:
        - key: HTTP_PORT
          value: {{ include "apps.webhooksApi.httpPort" . | quote }}

        - key: KL_HOOK_TRIGGER_AUTHZ_SECRET
          type: secret
          refName: {{ include "apps.webhooksApi.authenticationSecret.name" . }}
          refKey: KLOUDLITE_AUTHZ_SECRET

        - key: GITHUB_AUTHZ_SECRET
          type: secret
          refName: {{ include "apps.webhooksApi.authenticationSecret.name" . }}
          refKey: GITHUB_AUTHZ_SECRET

        - key: GITLAB_AUTHZ_SECRET
          type: secret
          refName: {{ include "apps.webhooksApi.authenticationSecret.name" . }}
          refKey: GITLAB_AUTHZ_SECRET

        - key: NATS_URL
          value: {{ include "nats.url" . }}

        - key: GIT_WEBHOOKS_TOPIC
          value: ""

        - key: COMMS_SERVICE
          value: 'comms:{{ include "apps.commsApi.httpPort" . }}'

        - key: DISCORD_WEBHOOK_URL
          value: "{{.Values.apps.webhooksApi.discordWebhookUrl}}"

        - key: WEBHOOK_URL
          value: "https://webhooks.{{.Values.webHost}}"

        - key: WEBHOOK_TOKEN_HASHING_SECRET
          type: secret
          refName: {{ include "apps.webhooksApi.authenticationSecret.name" . }}
          refKey: {{ include "apps.webhooksApi.authenticationSecret.token-key" . }}
