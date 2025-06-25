---
apiVersion: crds.kloudlite.io/v1
kind: Router
metadata:
  name: webhooks
  namespace: {{.Release.Namespace}}
spec:
  ingressClass: {{ .Values.ingress.ingressClass }}
  domains:
    - webhooks.{{.Values.webHost}}
  https:
    enabled: true
    forceRedirect: true
  routes:
    - app: webhooks-api
      path: /
      port: {{ include "apps.webhooksApi.httpPort" . }}
---
