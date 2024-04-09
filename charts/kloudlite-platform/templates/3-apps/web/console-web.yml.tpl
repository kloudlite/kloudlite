{{- if .Values.apps.consoleWeb.enabled}}
{{- $appName := "console-web" }}

apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: console-web
  namespace: {{.Release.Namespace}}
spec:
  serviceAccount: {{.Values.global.normalSvcAccount}}

  nodeSelector: {{include "stateless-node-selector" . | nindent 4 }}
  tolerations: {{include "stateless-tolerations" . | nindent 4 }}
  
  topologySpreadConstraints:
    {{ include "tsc-hostname" (dict "kloudlite.io/app.name" $appName) | nindent 4 }}
    {{ include "tsc-nodepool" (dict "kloudlite.io/app.name" $appName) | nindent 4 }}

  replicas: {{.Values.apps.consoleWeb.configuration.replicas}}

  services:
    - port: 80
      targetPort: 3000
      name: http
      type: tcp
  containers:
    - name: main
      image: {{.Values.apps.consoleWeb.image.repository}}:{{.Values.apps.consoleWeb.image.tag | default (include "image-tag" .) }}
      imagePullPolicy: {{ include "image-pull-policy" .}}
      resourceCpu:
        min: "100m"
        max: "300m"
      resourceMemory:
        min: "300Mi"
        max: "500Mi"
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
        - key: COOKIE_DOMAIN
          value: "{{.Values.global.cookieDomain}}"
        - key: GATEWAY_URL
          value: "http://gateway"
        - key: PORT
          value: "3000"
        - key: ARTIFACTHUB_KEY_ID
          value: {{.Values.apps.consoleWeb.configuration.artifactHubKeyID}}
        - key: ARTIFACTHUB_KEY_SECRET
          value: {{.Values.apps.consoleWeb.configuration.artifactHubKeySecret}}
        - key: GITHUB_APP_NAME
          type: secret
          refName: {{.Values.oAuth.secretName}}
          refKey: "GITHUB_APP_NAME"
---
{{- end}}
