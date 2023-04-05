---
apiVersion: crds.kloudlite.io/v1
kind: Router
metadata:
  name: {{.Values.routers.webhooksApi.name}}
  namespace: {{.Release.Namespace}}
  labels:
    kloudlite.io/account-ref: {{.Values.accountName}}
spec:
  region: {{.Values.region}}
  domains:
    - {{.Values.routers.webhooksApi.domain}}
  https:
    enabled: true
  routes:
    - app: {{.Values.apps.webhooksApi.name}}
      path: /
      port: 80
---
