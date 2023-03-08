---
apiVersion: crds.kloudlite.io/v1
kind: Router
metadata:
  name: {{.Values.routers.gatewayApi.name}}
  namespace: {{.Release.Namespace}}
  labels:
    kloudlite.io/account-ref: {{.Values.accountName}}
spec:
  domains:
    - {{.Values.routers.gatewayApi.domain}}
  https:
    enabled: true
    forceRedirect: true
  cors:
    enabled: true
    origins:
      - https://studio.apollographql.com
    allowCredentials: true
  basicAuth:
    enabled: true
    username: {{.Values.apps.gatewayApi.name}}
  routes:
    - app: {{.Values.apps.gatewayApi.name}}
      path: /
      port: 80
