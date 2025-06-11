{{- if .Values.apps.healthApi.install }}
{{- $appName := include "apps.healthApi.name" . }}

apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: "{{$appName}}"
  namespace: {{.Release.Namespace}}
  labels: {{ include "common.pod-labels" . | nindent 4}}
  annotations:
    {{ include "common.pod-annotations" . | nindent 4}}
spec:
  serviceAccount: {{.Values.serviceAccounts.clusterAdmin.name}}

  nodeSelector: {{.Values.scheduling.stateless.nodeSelector | toYaml | nindent 4}}
  tolerations: {{.Values.scheduling.stateless.tolerations | toYaml | nindent 4}}
  
  topologySpreadConstraints:
    {{ include "tsc-hostname" (dict "kloudlite.io/app.name" $appName) | nindent 4 }}

  replicas: {{.Values.apps.healthApi.replicas}}

  hpa:
    enabled: {{.Values.apps.healthApi.hpa.enabled}}
    minReplicas: {{.Values.apps.healthApi.hpa.minReplicas}}
    maxReplicas: {{.Values.apps.healthApi.hpa.maxReplicas}}
    thresholdCpu: 70
    thresholdMemory: 80

  services:
    - port: {{ include "apps.healthApi.httpPort" . }}

  containers:
    - name: main
      image: '{{.Values.apps.healthApi.image.repository}}:{{.Values.apps.healthApi.image.tag | default (include "image-tag" .) }}'
      imagePullPolicy: {{ include "image-pull-policy" .}}
      resourceCpu:
        min: "10m"
        max: "20m"
      resourceMemory:
        min: "10Mi"
        max: "20Mi"
      env:
        - key: HTTP_PORT
          value: {{ include "apps.healthApi.httpPort" . | squote}}

      livenessProbe: &probe
        type: httpGet
        httpGet:
          path: /_healthy
          port: {{ include "apps.healthApi.httpPort" . }}
        initialDelay: 5
        interval: 10

      readinessProbe: *probe
{{- end }}
