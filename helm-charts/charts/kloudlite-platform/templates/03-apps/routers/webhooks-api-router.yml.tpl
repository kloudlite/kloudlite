---
apiVersion: crds.kloudlite.io/v1
kind: Router
metadata:
  name: webhooks
  namespace: {{.Release.Namespace}}
spec:
  ingressClass: {{ .Values.nginxIngress.ingressClass.name }}
  domains:
    - webhooks.{{.Values.baseDomain}}
  https:
    enabled: true
    forceRedirect: true
  routes:
    - app: webhooks-api
      path: /
      port: {{ include "apps.webhooksApi.httpPort" . }}
---
