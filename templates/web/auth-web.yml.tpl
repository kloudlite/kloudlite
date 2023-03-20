apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: {{.Values.apps.authWeb.name}}
  namespace: {{.Release.Namespace}}
  labels:
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
      name: http
      type: tcp
  containers:
    - name: main
      image: {{.Values.apps.authWeb.image}}
      imagePullPolicy: {{.Values.apps.authWeb.ImagePullPolicy | default .Values.imagePullPolicy }}
      resourceCpu:
        min: "200m"
        max: "300m"
      resourceMemory:
        min: "200Mi"
        max: "300Mi"
      env:
        - key: BASE_URL
          value: "{{.Values.baseDomain}}"
        - key: ENV
          value: "{{.Values.envName}}"
        - key: PORT
          value: "3000"
---
