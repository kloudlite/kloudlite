---
apiVersion: crds.kloudlite.io/v1
kind: Router
metadata:
  name: {{.Values.routers.authWeb.name}}
  namespace: {{.Release.Namespace}}
spec:
  ingressClass: {{ (index .Values.helmCharts "ingress-nginx").configuration.ingressClassName }}
  domains:
    - {{.Values.routers.authWeb.name}}.{{.Values.baseDomain}}
  https:
    enabled: true
    clusterIssuer: {{.Values.clusterIssuer.name}}
    forceRedirect: true
  routes:
    - app: {{.Values.apps.authWeb.name}}
      path: /socket
      port: 6000
    - app: {{.Values.apps.authWeb.name}}
      path: /
      port: 80
---
