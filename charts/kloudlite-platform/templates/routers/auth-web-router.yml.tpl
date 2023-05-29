---
apiVersion: crds.kloudlite.io/v1
kind: Router
metadata:
  name: {{.Values.routers.authWeb.name}}
  namespace: {{.Release.Namespace}}
  labels:
    
spec:
  ingressClass: {{.Values.ingressClassName}}
  region: {{.Values.region}}
  domains:
    - {{.Values.routers.authWeb.domain}}
  https:
    enabled: true
    clusterIssuer: {{.Values.clusterIssuer.name}}
    forceRedirect: true
  routes:
    - app: {{.Values.apps.authWeb.name}}
      path: /
      port: 80
---
