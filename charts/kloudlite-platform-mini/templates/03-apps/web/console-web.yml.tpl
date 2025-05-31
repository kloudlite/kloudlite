{{- $appName :=  include "apps.consoleWeb.name" . }}

apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: {{ $appName }}
  namespace: {{.Release.Namespace}}
  labels: {{ include "common.pod-labels" . | nindent 4 }}
  annotations:
    kloudlite.io/checksum.supergraph: {{ include (print $.Template.BasePath "/03-apps/apis/secrets/gateway-supergraph.yml.tpl") . | sha256sum }}
    {{ include "common.pod-annotations" . | nindent 4 }}
spec:
  serviceAccount: {{.Values.serviceAccounts.normal.name}}

  nodeSelector: {{.Values.scheduling.stateless.nodeSelector | toYaml | nindent 4}}
  tolerations: {{.Values.scheduling.stateless.tolerations | toYaml | nindent 4}}
  topologySpreadConstraints:
    {{ include "tsc-hostname" (dict "kloudlite.io/app.name" $appName) | nindent 4 }}

  replicas: {{.Values.apps.consoleWeb.replicas}}

  services:
    - port: {{ include "apps.consoleWeb.httpPort" . }}
  
  router:
    routes:
      - host: "console.{{.Values.webHost}}"
        path: /
        port: {{ include "apps.consoleWeb.httpPort" . }}
        rewrite: false
        service: {{$appName}}

  containers:
    - name: main
      image: '{{.Values.apps.consoleWeb.image.repository}}:{{.Values.apps.consoleWeb.image.tag | default (include "image-tag" .) }}'
      imagePullPolicy: {{ include "image-pull-policy" .}}
      resourceCpu:
        min: "200m"
        max: "400m"
      resourceMemory:
        min: "300Mi"
        max: "500Mi"
      livenessProbe: &probe
        type: httpGet
        initialDelay: 5
        failureThreshold: 3
        httpGet:
          path: /healthy.txt
          port: {{ include "apps.consoleWeb.httpPort" . }}
        interval: 10
      readinessProbe: *probe
      env:
        - key: PORT
          value: {{ include "apps.consoleWeb.httpPort" . | quote }}
        - key: BASE_URL
          value: {{.Values.webHost}}
        - key: COOKIE_DOMAIN
          value: "{{- include "kloudlite.cookie-domain" . }}"
        - key: GATEWAY_URL
          value: 'http://{{ include "apps.gatewayApi.name" . }}:{{ include "apps.gatewayApi.httpPort" . }}'
        - key: ARTIFACTHUB_KEY_ID
          value: {{.Values.apps.consoleWeb.artifactHubKeyID}}
        - key: ARTIFACTHUB_KEY_SECRET
          value: {{.Values.apps.consoleWeb.artifactHubKeySecret}}
        - key: GITHUB_APP_NAME
          type: secret
          refName: {{ include "apps.authApi.oAuth2-secret.name" .}}
          refKey: "GITHUB_APP_NAME"
          optional: true
---
