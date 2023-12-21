apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: gateway
  namespace: {{.Release.Namespace}}
  annotations:
    config-checksum: {{ include (print $.Template.BasePath "/3-apps/apis/secrets/gateway-supergraph.yml.tpl") . | sha256sum }}
spec:
  serviceAccount: {{.Values.global.normalSvcAccount}}

  {{ include "node-selector-and-tolerations" . | nindent 2 }}
  
  services:
    - port: 80
      targetPort: 3000
      type: tcp
      name: http
  containers:
    - name: main
      image: {{.Values.apps.gatewayApi.image}}
      imagePullPolicy: {{.Values.global.imagePullPolicy }}
      {{if .Values.global.isDev}}
      args:
       - --dev
      {{end}}
      env:
        - key: PORT
          value: '3000'
        - key: SUPERGRAPH_CONFIG
          value: /kloudlite/config
      resourceCpu:
        min: 80m
        max: 200m
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
          port: 3000
        initialDelay: 20
        interval: 10

      readinessProbe:
        type: httpGet
        httpGet:
          path: /healthz
          port: 3000
        initialDelay: 20
        interval: 10
