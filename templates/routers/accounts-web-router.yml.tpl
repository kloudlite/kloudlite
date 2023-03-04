---
apiVersion: crds.kloudlite.io/v1
kind: Router
metadata:
  name: {{.Values.routers.accountsWeb.name}}
  namespace: {{.Release.Namespace}}
  labels:
    kloudlite.io/account-ref: {{.Values.accountName}}
spec:
  domains:
    - {{.Values.routers.accountsWeb.domain}}
  https:
    enabled: true
    forceRedirect: true
  routes:
    - app: {{.Values.apps.accountsWeb.name}}
      path: /
      port: 80
---
