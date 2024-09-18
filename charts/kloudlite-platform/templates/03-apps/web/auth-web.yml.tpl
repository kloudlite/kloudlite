{{- $appName := include "apps.authWeb.name" . -}}

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

  replicas: {{.Values.apps.authWeb.replicas}}
  
  services:
    - port: {{ include "apps.authWeb.httpPort" . }}

  containers:
    - name: main
      image: '{{.Values.apps.authWeb.image.repository}}:{{.Values.apps.authWeb.image.tag | default (include "image-tag" .) }}'
      imagePullPolicy: {{ include "image-pull-policy" .}}

      resourceCpu:
        min: "200m"
        max: "400m"
      resourceMemory:
        min: "200Mi"
        max: "400Mi"
      livenessProbe: &probe
        type: httpGet
        initialDelay: 5
        failureThreshold: 3
        httpGet:
          path: /healthy.txt
          port: {{ include "apps.authWeb.httpPort" . }}
        interval: 10
      readinessProbe: *probe
      env:
        - key: BASE_URL
          value: {{.Values.baseDomain}}
        - key: GATEWAY_URL
          value: 'http://{{ include "apps.gatewayApi.name" . }}:{{ include "apps.gatewayApi.httpPort" . }}'
        - key: COOKIE_DOMAIN
          value: "{{- include "kloudlite.cookie-domain" . }}"
        - key: PORT
          value: {{ include "apps.authWeb.httpPort" . | quote }}

