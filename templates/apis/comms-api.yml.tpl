apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: {{.Values.apps.commsApi.name}}
  namespace: {{.Release.Namespace}}
  annotations:
    kloudlite.io/account-ref: {{.Values.accountName}}
spec:
  accountName: {{.Values.accountName}}
  region: {{.Values.region}}

  {{ if .Values.nodeSelector }}
  nodeSelector: {{.Values.nodeSelector | toYaml | nindent 4}}
  {{ end }}
  
  {{- if .Values.tolerations -}}
  tolerations: {{.Values.tolerations | toYaml | nindent 4}}
  {{- end -}}

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
        min: "100m"
        max: "200m"
      resourceMemory:
        min: "100Mi"
        max: "200Mi"

      env:
        - key: GRPC_PORT
          value: "3001"

        - key: SUPPORT_EMAIL
          value: {{.Values.supportEmail}}

        - key: SENDGRID_API_KEY
          value: {{.Values.sendgridApiKey}}

        - key: EMAIL_LINKS_BASE_URL
          value: https://auth.{{.Values.baseDomain}}/

      # envFrom:
      #   - type: secret
      #     refName: comms-env
