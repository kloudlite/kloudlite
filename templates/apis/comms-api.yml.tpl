apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: {{.Values.apps.commsApi.name}}
  namespace: {{.Release.Namespace}}
  annotations:
    kloudlite.io/account-ref: {{.Values.accountName}}
spec:
  region: {{ .Values.region | default "" }}

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
      image: {{.Values.apps.commsApi.image}}
      imagePullPolicy: {{.Values.apps.commsApi.ImagePullPolicy | default .Values.imagePullPolicy }}
      resourceCpu:
        min: "50m"
        max: "80m"
      resourceMemory:
        min: "80Mi"
        max: "120Mi"

      env:
        - key: GRPC_PORT
          value: "3001"

        - key: SUPPORT_EMAIL
          value: {{.Values.supportEmail}}

        - key: SENDGRID_API_KEY
          value: {{.Values.sendgridApiKey}}

        - key: EMAIL_LINKS_BASE_URL
          value: https://auth.{{.Values.baseDomain}}/
