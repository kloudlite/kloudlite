{{/* [common tpl helpers] */}}

{{- define "pod-labels" -}}
{{- (.Values.global.podLabels | default dict) | toYaml -}}
{{- end -}}

{{- define "node-selector" -}}
{{- (.Values.global.nodeSelector | default dict) | toYaml -}}
{{- end -}}

{{- define "tolerations" -}}
{{- (.Values.global.tolerations | default list) | toYaml -}}
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

{{- define "is-required" -}}
{{- $field := index . 0 -}}
{{- $errorMsg := index . 1 -}}

{{- if not $field }}
{{ fail $errorMsg }}
{{- end }}

{{- end -}}

{{- define "preferred-node-affinity-to-masters" -}}
preferredDuringSchedulingIgnoredDuringExecution:
  - weight: 1
    preference:
      matchExpressions:
        - key: node-role.kubernetes.io/master
          operator: In
          values:
            - "true"
{{- end }}

{{- define "required-node-affinity-to-masters" -}}
requiredDuringSchedulingIgnoredDuringExecution:
  nodeSelectorTerms:
    - matchExpressions:
        - key: node-role.kubernetes.io/master
          operator: In
          values:
            - "true"
{{- end }}

{{- define "masters-node-selector" -}}
node-role.kubernetes.io/master: "true"
{{- end -}}

{{- define "masters-tolerations" -}}
- key: node-role.kubernetes.io/master
  operator: Exists
{{- end -}}

{{- define "router-domain" -}}
{{ .Values.global.routerDomain | default .Values.global.baseDomain }}
{{- end -}}

{{- define "image-tag" -}}
{{ .Values.global.kloudlite_release | default .Chart.AppVersion }}
{{- end -}}

{{- define "image-pull-policy" -}}
{{- if .Values.global.imagePullPolicy -}}
{{- .Values.global.imagePullPolicy}}
{{- else -}}
{{- if hasSuffix "-nightly" (include "image-tag" .) -}}
{{- "Always" }}
{{- else -}}
{{- "IfNotPresent" }}
{{- end -}}
{{- end -}}
{{- end -}}

{{- define "has-aws-vpc" -}}
{{ and .Values.operators.platformOperator.configuration.nodepools.aws.vpc_params.readFromCluster (eq .Values.operators.platformOperator.configuration.nodepools.cloudproviderName "aws") }}
{{- end -}}

{{- define "stateless-node-selector" -}}
{{.Values.nodepools.stateless.labels | toYaml}}
{{- end -}}

{{- define "stateful-node-selector" -}}
{{.Values.nodepools.stateful.labels | toYaml}}
{{- end -}}

{{- define "iac-node-selector" -}}
{{.Values.nodepools.iac.labels | toYaml}}
{{- end -}}

{{- define "stateless-tolerations" -}}
{{.Values.nodepools.stateless.tolerations | toYaml}}
{{- end -}}

{{- define "stateful-tolerations" -}}
{{.Values.nodepools.stateful.tolerations | toYaml}}
{{- end -}}

{{- define "iac-tolerations" -}}
{{.Values.nodepools.iac.tolerations | toYaml}}
{{- end -}}

{{- define "tsc-nodepool" -}}
- maxSkew: 1
  topologyKey: kloudlite.io/nodepool.name
  whenUnsatisfiable: DoNotSchedule
  nodeAffinityPolicy: Honor
  nodeTaintsPolicy: Honor
  labelSelector:
    matchLabels: {{ . | toYaml | nindent 6 }}
{{- end -}}

{{- define "tsc-hostname" -}}
- maxSkew: 1
  topologyKey: kubernetes.io/hostname
  whenUnsatisfiable: DoNotSchedule
  nodeAffinityPolicy: Honor
  nodeTaintsPolicy: Honor
  labelSelector:
    matchLabels: {{ . | toYaml | nindent 6 }}
{{- end -}}

{{- define "prom-http-addr" -}}
http://vmselect-{{ $.Values.victoriaMetrics.name }}.{{$.Release.Namespace}}.svc.{{$.Values.global.clusterInternalDNS}}:8481/select/0/prometheus
{{- end -}}
