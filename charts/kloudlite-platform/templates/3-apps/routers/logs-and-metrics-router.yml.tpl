---
apiVersion: crds.kloudlite.io/v1
kind: Router
metadata:
  name: observe
  namespace: {{.Release.Namespace}}
spec:
  ingressClass: {{ .Values.global.ingressClassName }}

  domains:
    - "observe.{{.Values.global.baseDomain}}"
  https:
    enabled: true
    forceRedirect: true
  routes:
    - app: console-api
      path: /
      port: 9100
---