{{- if .Values.certManager.clusterIssuer.create }}
---
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: {{.Values.certManager.clusterIssuer.name}}
spec:
  acme:
    email: {{.Values.certManager.clusterIssuer.acmeEmail}}
    privateKeySecretRef:
      name: {{.Values.certManager.clusterIssuer.name}}
    server: https://acme-v02.api.letsencrypt.org/directory
    solvers:
      {{- if .Values.certManager.solvers }}
      {{- range $v := .Values.certManager.solvers }}
      - {{$v}}
      {{- end }}
      {{- else }}
      - http01:
          ingress:
            class: {{.Values.nginxIngress.ingressClass.name | squote}}
      {{- end }}
{{- end }}
