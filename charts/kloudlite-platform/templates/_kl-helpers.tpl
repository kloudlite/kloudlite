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

{{- define "observability-annotations-resource" -}}

{{- $resourceName := index . 0 }}
{{- $resourceType := index . 1 }}
{{- $resourceComponent := "" }}
{{- if gt (len .) 2 }}
{{- $resourceComponent := index . 2 }}
{{- end }}

kloudlite.io/resource_name: {{$resourceName}}
kloudlite.io/resource_type: {{$resourceType}}
{{- if $resourceComponent }}
kloudlite.io/resource_component: {{$resourceComponent}}
{{- end }}

{{- end }}
{{- /**/}}
{{- /* {{- if $workspaceName }} */}}
{{- /* kloudlite.io/workspace_name: "{{$workspaceName}}" */}}
{{- /* {{- end }} */}}
{{- /**/}}
{{- /* {{- if $workspaceTargetNs }} */}}
{{- /* kloudlite.io/workspace_target_ns: "{{$workspaceTargetNs}}" */}}
{{- /* {{- end }} */}}
{{- /**/}}
{{- /* {{- if $projectName}} */}}
{{- /* kloudlite.io/project_name: "{{$projectName}}" */}}
{{- /* {{- end }} */}}
{{- /**/}}
{{- /* {{- if $projectTargetNs}} */}}
{{- /* kloudlite.io/project_target_ns: "{{$projectTargetNs}}" */}}
{{- /* {{- end }} */}}
{{- /**/}}
{{- /* {{- end -}} */}}
{{- /* {{- end -}} */}}
