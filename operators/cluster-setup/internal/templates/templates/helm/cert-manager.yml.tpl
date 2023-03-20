{{ $podLabels := get . "pod-labels" |  default dict }} 
{{ $tolerations := get . "tolerations" | default list }}
{{ $nodeSelector := get . "node-selector" | default dict }}

installCRDs: true

extraArgs:
  - "--dns01-recursive-nameservers-only"
  - "--dns01-recursive-nameservers=1.1.1.1:53,8.8.8.8:53"

{{- if $tolerations }}
tolerations: {{ $tolerations | mustFromJson | toYAML | nindent 2}}
{{- end }}

{{- if $podLabels }}
podLabels: {{$podLabels | mustFromJson | toYAML | nindent 2}}
{{- end }}

resources:
  limits:
    cpu: 80m
    memory: 120Mi
  requests:
    cpu: 40m
    memory: 120Mi

webhook:
  {{- if $podLabels }}
  podLabels: {{$podLabels | mustFromJson | toYAML | nindent 4}}
  {{- end }}
  resources:
    limits:
      cpu: 60m
      memory: 60Mi
    requests:
      cpu: 30m
      memory: 60Mi

cainjector:
  {{- if $podLabels }}
  podLabels: {{$podLabels | mustFromJson | toYAML | nindent 4}}
  {{- end }}
  resources:
    limits:
      cpu: 120m
      memory: 200Mi
    requests:
      cpu: 80m
      memory: 200Mi
