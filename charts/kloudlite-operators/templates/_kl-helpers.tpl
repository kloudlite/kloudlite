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
