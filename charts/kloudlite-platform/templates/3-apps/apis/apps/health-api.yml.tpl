{{- $appName := "health-api" }}

apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: health-api
  namespace: {{.Release.Namespace}}
spec:
  serviceAccount: {{.Values.global.clusterSvcAccount}}

  nodeSelector: {{include "stateless-node-selector" . | nindent 4 }}
  tolerations: {{include "stateless-tolerations" . | nindent 4 }}
  
  topologySpreadConstraints:
    {{ include "tsc-hostname" (dict "kloudlite.io/app.name" $appName) | nindent 4 }}
    {{ include "tsc-nodepool" (dict "kloudlite.io/app.name" $appName) | nindent 4 }}

  replicas: {{.Values.apps.healthApi.configuration.replicas}}

  services:
    - port: {{.Values.apps.healthApi.configuration.httpPort | int }}

  containers:
    - name: main
      image: {{.Values.apps.healthApi.image.repository}}:{{.Values.apps.healthApi.image.tag | default (include "image-tag" .) }}
      imagePullPolicy: {{ include "image-pull-policy" .}}
      resourceCpu:
        min: "10m"
        max: "20m"
      resourceMemory:
        min: "10Mi"
        max: "20Mi"
      env:
        - key: HTTP_PORT
          value: {{.Values.apps.healthApi.configuration.httpPort | squote}}

      livenessProbe:
        type: httpGet
        httpGet:
          path: /_healthy
          port: {{.Values.apps.healthApi.configuration.httpPort}}
        initialDelay: 5
        interval: 10

      readinessProbe:
        type: httpGet
        httpGet:
          path: /_healthy
          port: {{.Values.apps.healthApi.configuration.httpPort}}
        initialDelay: 5
        interval: 10
