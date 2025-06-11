---
apiVersion: crds.kloudlite.io/v1
kind: Router
metadata:
  name: auth
  namespace: {{.Release.Namespace}}
spec:
  ingressClass: {{ .Values.ingress.ingressClass }}
  domains:
    - auth.{{.Values.webHost}}
  https:
    enabled: true
    forceRedirect: true
  routes:
    - app: infra-api
      path: /render/helm
      port: {{  include "apps.infraApi.httpPort" . }}
      rewrite: false

    - app: auth-web
      path: /
      port: {{ include "apps.authWeb.httpPort" . }}

---
