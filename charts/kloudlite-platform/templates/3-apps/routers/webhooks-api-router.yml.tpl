---
apiVersion: crds.kloudlite.io/v1
kind: Router
metadata:
  name: webhook
  namespace: {{.Release.Namespace}}
spec:
  ingressClass: {{.Values.global.ingressClassName}}
  domains:
    - "webhook.{{.Values.global.baseDomain}}"
  https:
    enabled: true
    forceRedirect: true
  routes:
    - app: webhooks-api
      path: /
      port: 80
---
