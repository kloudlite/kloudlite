{{- define "serviceAccountName" -}}
{{- printf "%s-%s" .Release.Name .Values.svcAccountName -}}
{{- end}}

{{- define "pod-labels" -}}
{{- (.Values.podLabels | default dict) | toYaml -}}
{{- end -}}

{{- define "node-selector" -}}
{{- (.Values.nodeSelector | default dict) | toYaml -}}
{{- end -}}

{{- define "tolerations" -}}
{{- (.Values.tolerations | default list) | toYaml -}}
{{- end -}}


{{- define "required-node-affinity-to-masters" -}}
requiredDuringSchedulingIgnoredDuringExecution:
  nodeSelectorTerms:
    - matchExpressions:
        - key: node-role.kubernetes.io/master
          operator: In
          values:
            - "true"
{{- end }}

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

{{- define "image-tag" -}}
{{ .Values.kloudliteRelease | default .Chart.AppVersion }}
{{- end -}}

{{- define "image-pull-policy" -}}
{{- hasSuffix "-nightly" . | ternary "Always" "IfNotPresent"}}
{{- end -}}

{{- /* {{- define "msvc-creds-svc-shared-secret" -}} */}}
{{- /* {{- if $.Values.operators.agentOperator.configuration.msvc.credsSvc.sharedSecret -}} */}}
{{- /* {{$.Values.operators.agentOperator.configuration.msvc.credsSvc.sharedSecret}} */}}
{{- /* {{- else -}} */}}
{{- /* {{ printf "%s/%s" $.Values.accountName $.Values.clusterName | sha256sum }} */}}
{{- /* {{- end -}} */}}
{{- /* {{- end -}} */}}
