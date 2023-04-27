---
apiVersion: crds.kloudlite.io/v1
kind: Router
metadata:
  name: {{.Values.routers.gatewayApi.name}}
  namespace: {{.Release.Namespace}}
  labels:
    kloudlite.io/account-ref: {{.Values.accountName}}
spec:
  ingressClass: {{.Values.ingressClassName}}
  region: {{.Values.region}}
  domains:
    - {{.Values.routers.gatewayApi.domain}}
  https:
    enabled: true
    clusterIssuer: {{.Values.clusterIssuer.name}}
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
