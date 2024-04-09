{{- if .Values.apps.authWeb.enabled -}}
{{- $appName := "auth-web" }}

apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: auth-web
  namespace: {{.Release.Namespace}}
spec:
  serviceAccount: {{.Values.global.normalSvcAccount}}

  nodeSelector: {{include "stateless-node-selector" . | nindent 4 }}
  tolerations: {{include "stateless-tolerations" . | nindent 4 }}
  
  topologySpreadConstraints:
    {{ include "tsc-hostname" (dict "kloudlite.io/app.name" $appName) | nindent 4 }}
    {{ include "tsc-nodepool" (dict "kloudlite.io/app.name" $appName) | nindent 4 }}

  replicas: {{.Values.apps.authWeb.configuration.replicas}}
  
  services:
    - port: 80
      targetPort: 3000
      name: http
      type: tcp
    - port: 6000
      targetPort: 6000
      name: http
      type: tcp
  containers:
    - name: main
      image: {{.Values.apps.authWeb.image.repository}}:{{.Values.apps.authWeb.image.tag | default (include "image-tag" .) }}
      imagePullPolicy: {{ include "image-pull-policy" .}}

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
          value: {{include "router-domain" .}}
        - key: GATEWAY_URL
          value: "http://gateway"
        - key: COOKIE_DOMAIN
          value: "{{.Values.global.cookieDomain}}"
        - key: PORT
          value: "3000"
---
{{- end -}}
