{{- $certManagerValues := get . "cert-manager-values" -}}

installCRDs: true

extraArgs:
  - "--dns01-recursive-nameservers-only"
  - "--dns01-recursive-nameservers=1.1.1.1:53,8.8.8.8:53"

{{- with $certManagerValues}}
{{/*gotype: github.com/kloudlite/operator/apis/cluster-setup/v1.CertManagerValues*/}}
{{- if .Tolerations}}
tolerations: {{ .Tolerations | toYAML | nindent 2 }}
{{- end }}

{{- if .PodLabels }}
podLabels: {{ .PodLabels | toYAML | nindent 2 }}
{{- end }}

{{- if .NodeSelector}}
nodeSelector: {{ .NodeSelector | toYAML | nindent 2 }}
{{- end }}

resources:
  limits:
    cpu: 80m
    memory: 120Mi
  requests:
    cpu: 40m
    memory: 120Mi

webhook:
  {{- if .PodLabels }}
  podLabels: {{ .PodLabels | toYAML | nindent 4}}
  {{- end }}
  resources:
    limits:
      cpu: 60m
      memory: 60Mi
    requests:
      cpu: 30m
      memory: 60Mi

cainjector:
  {{- if .PodLabels }}
  podLabels: {{ .PodLabels | toYAML | nindent 4 }}
  {{- end }}
  resources:
    limits:
      cpu: 120m
      memory: 200Mi
    requests:
      cpu: 80m
      memory: 200Mi
{{- end }}
