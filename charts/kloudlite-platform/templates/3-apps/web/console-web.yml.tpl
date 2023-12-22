{{- if .Values.apps.consoleWeb.enabled}}
apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: console-web
  namespace: {{.Release.Namespace}}
spec:
  serviceAccount: {{.Values.global.normalSvcAccount}}

  {{ include "node-selector-and-tolerations" . | nindent 2 }}

  services:
    - port: 80
      targetPort: 3000
      name: http
      type: tcp
  containers:
    - name: main
      image: {{.Values.apps.consoleWeb.image}}
      imagePullPolicy: {{.Values.global.imagePullPolicy }}
      resourceCpu:
        min: "100m"
        max: "200m"
      resourceMemory:
        min: "200Mi"
        max: "300Mi"
      livenessProbe: &probe
        type: httpGet
        initialDelay: 5
        failureThreshold: 3
        httpGet:
          path: /healthy.txt
          port: 3000
        interval: 10
      readinessProbe: *probe
      env:
        - key: BASE_URL
          value: "{{.Values.global.baseDomain}}"
        - key: COOKIE_DOMAIN
          value: "{{.Values.global.cookieDomain}}"
        - key: GATEWAY_URL
          value: "http://gateway"
        - key: PORT
          value: "3000"
---
{{- end}}
