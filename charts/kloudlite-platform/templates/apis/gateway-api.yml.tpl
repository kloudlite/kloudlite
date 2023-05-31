apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: {{.Values.apps.gatewayApi.name}}
  namespace: {{.Release.Namespace}}
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
          value: /hotspot/config
      resourceCpu:
        min: 80m
        max: 200m
      resourceMemory:
        min: 200Mi
        max: 300Mi

      volumes:
        - mountPath: /hotspot
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

---
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{.Values.apps.gatewayApi.name}}-supergraph
  namespace: {{.Release.Namespace}}
data:
  config: |+
    serviceList:
      - name: {{.Values.apps.authApi.name}}
        url: http://{{.Values.apps.authApi.name}}.{{.Release.Namespace}}.svc.cluster.local/query

      - name: {{.Values.apps.containerRegistryApi.name}}
        url: http://{{.Values.apps.containerRegistryApi.name}}.{{.Release.Namespace}}.svc.cluster.local/query

      - name: {{.Values.apps.consoleApi.name}}
        url: http://{{.Values.apps.consoleApi.name}}.{{.Release.Namespace}}.svc.cluster.local/query

      - name: {{.Values.apps.financeApi.name}}
        url: http://{{.Values.apps.financeApi.name}}.{{.Release.Namespace}}.svc.cluster.local/query

      - name: {{.Values.apps.infraApi.name}}
        url: http://{{.Values.apps.infraApi.name}}.{{.Release.Namespace}}.svc.cluster.local/query

      - name: {{.Values.apps.messageOfficeApi.name}}
        url: http://{{.Values.apps.messageOfficeApi.name}}.{{.Release.Namespace}}.svc.cluster.local/query
---
