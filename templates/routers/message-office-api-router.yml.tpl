---
apiVersion: crds.kloudlite.io/v1
kind: Router
metadata:
  name: {{.Values.routers.messageOfficeApi.name}}
  namespace: {{.Release.Namespace}}
  labels:
    kloudlite.io/account-ref: {{.Values.accountName}}
spec:
  ingressClass: {{.Values.ingressClassName}}
  region: {{.Values.region}}
  backendProtocol: GRPC
  domains:
    - {{.Values.routers.messageOfficeApi.domain}}
  https:
    enabled: true
    clusterIssuer: {{.Values.clusterIssuer.name}}
    forceRedirect: true
  routes:
    - app: {{.Values.apps.messageOfficeApi.name}}
      path: /
      port: 3001
---
