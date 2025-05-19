{{- if .Values.apps.healthApi.install }}
---
apiVersion: crds.kloudlite.io/v1
kind: Router
metadata:
  name: health-api
  namespace: {{.Release.Namespace}}
spec:
  ingressClass: {{ .Values.ingress.ingressClass }}
  domains:
    - health.{{.Values.baseDomain}}
  https:
    enabled: true
    forceRedirect: true
  routes:
    - app: health-api
      path: /kubernetes
      port: {{ include "apps.healthApi.httpPort" . }}
      rewrite: false

    - app: health-api
      path: /{{.Release.Namespace}}/*
      port: {{ include "apps.healthApi.httpPort" . }}
      rewrite: false

    - app: health-api
      path: /sts/{{.Release.Namespace}}/*
      port: {{ include "apps.healthApi.httpPort" . }}
      rewrite: false

    - app: health-api
      path: /deploy/{{.Release.Namespace}}/*
      port: {{ include "apps.healthApi.httpPort" . }}
      rewrite: false
---
{{- end }}
