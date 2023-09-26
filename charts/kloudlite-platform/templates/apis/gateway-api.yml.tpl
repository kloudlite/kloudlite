apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: {{.Values.apps.gatewayApi.name}}
  namespace: {{.Release.Namespace}}
  annotations:
    config-checksum: {{ include (print $.Template.BasePath "/configmaps/gateway-supergraph.yml.tpl") . | sha256sum }}
spec:
  region: {{.Values.region | default ""}}
  serviceAccount: {{.Values.normalSvcAccount}}

  {{ include "node-selector-and-tolerations" . | nindent 2 }}
  
  services:
    - port: 80
      targetPort: 3000
      type: tcp
      name: http
  containers:
    - name: main
      image: {{.Values.apps.gatewayApi.image}}
      imagePullPolicy: {{.Values.apps.gatewayApi.ImagePullPolicy | default .Values.imagePullPolicy }}
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
          refName: {{.Values.apps.gatewayApi.name}}-supergraph

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
