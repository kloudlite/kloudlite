---
apiVersion: crds.kloudlite.io/v1
kind: Router
metadata:
  name: webhooks
  namespace: {{.Release.Namespace}}
spec:
  ingressClass: {{.Values.global.ingressClassName}}
  domains:
    - webhooks.{{include "router-domain" .}}
  https:
    enabled: true
    forceRedirect: true
  routes:
    - app: webhooks-api
      path: /
      port: 3000
---
