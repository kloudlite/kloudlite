{{- if .Values.apps.klInstaller.install }}
---
apiVersion: crds.kloudlite.io/v1
kind: Router
metadata:
  name: kl-installer
  namespace: {{.Release.Namespace}}
spec:
  ingressClass: {{ .Values.nginxIngress.ingressClass.name }}
  domains:
    - kl.{{.Values.baseDomain}}
  https:
    enabled: true
    forceRedirect: true
  routes:
    - app: kl-installer
      path: /
      port: 3000
      rewrite: false
---
{{- end }}
