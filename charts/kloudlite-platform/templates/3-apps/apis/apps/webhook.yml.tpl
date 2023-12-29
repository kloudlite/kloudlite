apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: webhooks-api
  namespace: {{.Release.Namespace}}
spec:
  serviceAccount: {{.Values.global.normalSvcAccount}}

  {{ include "node-selector-and-tolerations" . | nindent 2 }}

  services:
    - port: 80
      targetPort: 3001
      name: http
      type: tcp
  containers:
    - name: main
      image: {{.Values.apps.webhooksApi.image}}
      imagePullPolicy: {{.Values.global.imagePullPolicy }}
      {{if .Values.global.isDev}}
      args:
       - --dev
      {{end}}
      
      resourceCpu:
        min: "30m"
        max: "50m"
      resourceMemory:
        min: "50Mi"
        max: "100Mi"
      env:

        - key: HTTP_PORT
          value: "3001"

        - key: KL_HOOK_TRIGGER_AUTHZ_SECRET
          type: secret
          refName: {{.Values.webhookSecrets.name}}
          refKey: KLOUDLITE_AUTHZ_SECRET

        - key: GITHUB_AUTHZ_SECRET
          type: secret
          refName: {{.Values.webhookSecrets.name}}
          refKey: GITHUB_AUTHZ_SECRET

        - key: GITLAB_AUTHZ_SECRET
          type: secret
          refName: {{.Values.webhookSecrets.name}}
          refKey: GITLAB_AUTHZ_SECRET

        - key: NATS_URL
          value: "nats://nats:4222"

        - key: GIT_WEBHOOKS_TOPIC
          value: "{{.Values.global.cookieDomain}}"


