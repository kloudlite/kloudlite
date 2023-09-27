---
apiVersion: crds.kloudlite.io/v1
kind: Router
metadata:
  name: {{.Values.routers.webhooksApi.name}}
  namespace: {{.Release.Namespace}}
spec:
  ingressClass: {{ (index .Values.helmCharts "ingress-nginx").configuration.ingressClassName }}
  domains:
    - "{{.Values.routers.webhooksApi.name}}.{{.Values.baseDomain}}"
  https:
    enabled: true
    clusterIssuer: {{.Values.clusterIssuer.name}}
    forceRedirect: true
  routes:
    - app: {{.Values.apps.webhooksApi.name}}
      path: /
      port: 80
---
