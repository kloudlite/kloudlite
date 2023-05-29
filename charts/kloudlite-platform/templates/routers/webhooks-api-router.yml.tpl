---
apiVersion: crds.kloudlite.io/v1
kind: Router
metadata:
  name: {{.Values.routers.webhooksApi.name}}
  namespace: {{.Release.Namespace}}
  labels:
    
spec:
  ingressClass: {{.Values.ingressClassName}}
  region: {{.Values.region}}
  domains:
    - {{.Values.routers.webhooksApi.domain}}
  https:
    enabled: true
    clusterIssuer: {{.Values.clusterIssuer.name}}
    forceRedirect: true
  routes:
    - app: {{.Values.apps.webhooksApi.name}}
      path: /
      port: 80
---
