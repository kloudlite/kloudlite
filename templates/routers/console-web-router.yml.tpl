---
apiVersion: crds.kloudlite.io/v1
kind: Router
metadata:
  name: {{.Values.routers.consoleWeb.name}}
  namespace: {{.Release.Namespace}}
  labels:
    kloudlite.io/account-ref: {{.Values.accountName}}
spec:
  ingressClass: {{.Values.ingressClassName}}
  region: {{.Values.region}}
  domains:
    - {{.Values.routers.consoleWeb.domain}}
  https:
    enabled: true
    clusterIssuer: {{.Values.clusterIssuer.name}}
    forceRedirect: true
  routes:
    - app: {{.Values.apps.consoleWeb.name}}
      path: /
      port: 80
---
