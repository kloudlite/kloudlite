---
apiVersion: crds.kloudlite.io/v1
kind: Router
metadata:
  name: {{.Values.routers.authWeb.name}}
  namespace: {{.Release.Namespace}}
  labels:
    kloudlite.io/account-ref: {{.Values.accountName}}
spec:
  domains:
    - {{.Values.routers.authWeb.domain}}
  https:
    enabled: true
    forceRedirect: true
  routes:
    - app: {{.Values.apps.authWeb.name}}
      path: /
      port: 80
---
