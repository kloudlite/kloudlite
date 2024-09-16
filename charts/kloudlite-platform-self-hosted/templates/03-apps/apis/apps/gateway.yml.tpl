{{- $appName := "gateway-api" }}

apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: {{$appName}}
  namespace: {{.Release.Namespace}}
  labels: {{ include "common.pod-labels" . | nindent 4}}
  annotations:
    kloudlite.io/checksum.supergraph: {{ include (print $.Template.BasePath "/03-apps/apis/secrets/gateway-supergraph.yml.tpl") . | sha256sum }}
    {{ include "common.pod-annotations" .}}
spec:
  serviceAccount: {{.Values.serviceAccounts.clusterAdmin.name}}

  nodeSelector: {{.Values.scheduling.stateless.nodeSelector | toYaml | nindent 4}}
  tolerations: {{.Values.scheduling.stateless.tolerations | toYaml | nindent 4}}
  topologySpreadConstraints:
    {{ include "tsc-hostname" (dict "kloudlite.io/app.name" $appName) | nindent 4 }}

  replicas: {{.Values.apps.gatewayApi.replicas}}

  hpa:
    enabled: {{.Values.apps.gatewayApi.hpa.enabled}}
    minReplicas: {{.Values.apps.gatewayApi.minReplicas}}
    maxReplicas: {{.Values.apps.gatewayApi.maxReplicas}}
    thresholdCpu: 70
    thresholdMemory: 80

  services:
    - port: {{ include "apps.gatewayApi.httpPort" . }}
  containers:
    - name: main
      image: '{{.Values.apps.gatewayApi.image.repository}}:{{.Values.apps.gatewayApi.image.tag | default (include "image-tag" .) }}'
      imagePullPolicy: {{ include "image-pull-policy" .}}
      env:
        - key: PORT
          value: {{ include "apps.gatewayApi.httpPort" . | squote}}
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

      livenessProbe: &probe
        type: httpGet
        httpGet:
          path: /healthz 
          port: {{ include "apps.gatewayApi.httpPort" . }}
        initialDelay: 20
        interval: 10

      readinessProbe: *probe
