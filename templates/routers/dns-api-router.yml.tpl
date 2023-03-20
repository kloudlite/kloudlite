apiVersion: crds.kloudlite.io/v1
kind: Router
metadata:
  name: {{.Values.routers.dnsApi.name}}
  namespace: {{.Release.Namespace}}
spec:
  domains:
    - "{{.Values.routers.dnsApi.domain}}"
  https:
    enabled: true
    forceRedirect: true
  basicAuth:
    enabled: true
    username:  {{.Values.routers.dnsApi.name}}
  routes:
    - app: {{.Values.apps.dnsApi.name}}
      path: /
      port: 80
