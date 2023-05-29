{{- if .Values.apps.commsApi.enabled -}}

apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: {{ .Values.apps.commsApi.name }}
  namespace: {{.Release.Namespace}}
spec:
  region: {{ .Values.region | default "" }}

  {{ include "node-selector-and-tolerations" . | nindent 2 }}

  services:
    - port: {{.Values.apps.commsApi.configuration.grpcPort}}
      targetPort: {{.values.apps.commsApi.configuration.grpcPort}}
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
          value: {{.values.apps.commsApi.configuration.grpcPort | squote}}

        - key: SUPPORT_EMAIL
          value: {{.Values.apps.commsApi.configuration.supportEmail}}

        - key: SENDGRID_API_KEY
          value: {{.Values.apps.commsApi.configuration.sendgridApiKey}}
        
        {{/* TODO: url should definitely NOT be auth.{{.Values.baseDomain}} */}}
        - key: EMAIL_LINKS_BASE_URL
          value: https://auth.{{.Values.baseDomain}}/
{{- end -}}
