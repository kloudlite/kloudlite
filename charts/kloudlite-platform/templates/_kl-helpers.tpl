{{/* [common tpl helpers] */}}

{{- define "pod-labels" -}}
{{- (.Values.podLabels | default dict) | toYaml -}}
{{- end -}}

{{- define "node-selector" -}}
{{- (.Values.nodeSelector | default dict) | toYaml -}}
{{- end -}}

{{- define "tolerations" -}}
{{- (.Values.tolerations | default list) | toYaml -}}
{{- end -}}

{{- define "node-selector-and-tolerations" -}}
{{- if .Values.nodeSelector -}}
nodeSelector: {{ include "node-selector" . | nindent 2 }}
{{- end }}

{{- if .Values.tolerations -}}
tolerations: {{ include "tolerations" . | nindent 2 }}
{{- end -}}
{{- end -}}

{{/* # -- wildcard certificate */}}
{{- define "cloudflare-wildcard-certificate.secret-name" -}}
{{- printf "%s-tls" .Values.cloudflareWildCardCert.name }}
{{- end -}}


{{- define "build-router-domain" -}}
{{- $name := index . 0 -}}
{{- $baseDomain := index . 1 -}}
{{- printf "%s.%s" $name $baseDomain -}}
{{- end -}}

{{/* [helm: redpanda-operator] */}}
{{- define "redpanda-operator.name" -}}
{{- printf "%s-redpanda-operator" .Release.Name -}}
{{- end -}}

{{/* [helm: cert-manager] */}}
{{- define "cert-manager.name" -}}
{{- printf "%s-cert-manager" .Release.Name -}}
{{- end -}}

{{/* helm: ingress-nginx */}}
{{- define "ingress-nginx.name" -}}
{{- printf "%s-ingress-nginx" .Release.Name -}}
{{- end -}}

{{/* [helm: grafana] */}}
{{- define "grafana.name" -}}
{{- printf "%s-grafana" .Release.Name -}}
{{- end -}}

{{/* helm: vector */}}
{{- define "vector.name" -}}
{{- printf "%s-vector" .Release.Name -}}
{{- end -}}

{{/* helm: loki */}}
{{- define "loki.name" -}}
{{- printf "%s-loki" .Release.Name -}}
{{- end -}}

{{/* [helm: kube-prometheus] */}}
{{- define "kube-prometheus.name" -}}
{{- printf "%s-prometheus" .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

