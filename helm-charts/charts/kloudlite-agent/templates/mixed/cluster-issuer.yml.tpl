{{- $chartOpts := .Values.helmCharts.certManager }} 
{{- if $chartOpts.enabled }}

{{- range $v := $chartOpts.configuration.clusterIssuers }}
---
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: {{$v.name}}
  annotations:
    kloudlite.io/description: "created by kloudlite agent helm chart"
spec:
  acme:
    email: {{$v.acme.email}}
    privateKeySecretRef:
      name: {{$v.name}}
    server: {{$v.acme.server}}
    solvers:
      {{- $ingClass := $.Values.helmCharts.ingressNginx.configuration.ingressClassName }}
      {{- if $ingClass }}
      - http01:
          ingress:
            class: "{{$ingClass}}"
      {{- end}}

{{- end }}
{{- end }}
