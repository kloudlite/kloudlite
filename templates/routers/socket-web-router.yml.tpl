---
apiVersion: crds.kloudlite.io/v1
kind: Router
metadata:
  name: {{.Values.routers.socketWeb.name}}
  namespace: {{.Release.Namespace}}
  labels:
    kloudlite.io/account-ref: {{.Values.accountName}}
spec:
  domains:
    - {{.Values.routers.socketWeb.domain}}
  https:
    enabled: true
    forceRedirect: true
  routes:
    - app: {{.Values.apps.socketWeb.name}}
      path: /
      port: 80
    - app: {{.Values.apps.socketWeb.name}}
      path: /publish
      port: 3001
---
