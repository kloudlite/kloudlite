{{- $appName := "gateway" }}

apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: gateway
  namespace: {{.Release.Namespace}}
  annotations:
    config-checksum: {{ include (print $.Template.BasePath "/3-apps/apis/secrets/gateway-supergraph.yml.tpl") . | sha256sum }}
spec:
  serviceAccount: {{.Values.global.normalSvcAccount}}

  nodeSelector: {{include "stateless-node-selector" . | nindent 4 }}
  tolerations: {{include "stateless-tolerations" . | nindent 4 }}
  
  topologySpreadConstraints:
    {{ include "tsc-hostname" (dict "kloudlite.io/app.name" $appName) | nindent 4 }}
    {{ include "tsc-nodepool" (dict "kloudlite.io/app.name" $appName) | nindent 4 }}

  replicas: {{.Values.apps.gatewayApi.configuration.replicas}}

  services:
    - port: {{.Values.apps.gatewayApi.configuration.httpPort}}
  containers:
    - name: main
      image: {{.Values.apps.gatewayApi.image.repository}}:{{.Values.apps.gatewayApi.image.tag | default (include "image-tag" .) }}
      imagePullPolicy: {{ include "image-pull-policy" .}}
      {{if .Values.global.isDev}}
      args:
       - --dev
      {{end}}
      env:
        - key: PORT
          value: {{.Values.apps.gatewayApi.configuration.httpPort | squote}}
        - key: SUPERGRAPH_CONFIG
          value: /kloudlite/config
      resourceCpu:
        min: 100m
        max: 1000m
      resourceMemory:
        min: 200Mi
        max: 300Mi

      volumes:
        - mountPath: /kloudlite
          type: config
          refName: gateway-supergraph

      livenessProbe:
        type: httpGet
        httpGet:
          path: /healthz 
          port: {{.Values.apps.gatewayApi.configuration.httpPort}}
        initialDelay: 10
        interval: 10

      readinessProbe:
        type: httpGet
        httpGet:
          path: /healthz
          port: {{.Values.apps.gatewayApi.configuration.httpPort}}
        initialDelay: 7
        interval: 10
