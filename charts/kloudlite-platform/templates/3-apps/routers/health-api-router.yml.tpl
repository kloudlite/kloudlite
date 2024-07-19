---
apiVersion: crds.kloudlite.io/v1
kind: Router
metadata:
  name: health-api
  namespace: {{.Release.Namespace}}
spec:
  ingressClass: {{ .Values.global.ingressClassName }}
  domains:
    - health.{{include "router-domain" .}}
  https:
    enabled: true
    forceRedirect: true
  routes:
    - app: health-api
      path: /kubernetes
      port: {{.Values.apps.healthApi.configuration.httpPort | int }}
      rewrite: false
    - app: health-api
      path: /{{.Release.Namespace}}/*
      port:  {{.Values.apps.healthApi.configuration.httpPort | int }}
      rewrite: false
---
