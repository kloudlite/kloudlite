---
apiVersion: crds.kloudlite.io/v1
kind: Router
metadata:
  name: {{.Values.routers.messageOfficeApi.name}}
  namespace: {{.Release.Namespace}}
spec:
  ingressClass: {{ (index .Values.helmCharts "ingress-nginx").configuration.ingressClassName }}
  backendProtocol: GRPC
  maxBodySizeInMB: 50
  domains:
    - "{{.Values.routers.messageOfficeApi.name}}.{{.Values.baseDomain}}"
  https:
    enabled: true
    clusterIssuer: {{.Values.clusterIssuer.name}}
    forceRedirect: true
  routes:
    - app: {{.Values.apps.messageOfficeApi.name}}
      path: /
      port: 3001
---
