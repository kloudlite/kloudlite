{{- define "service-account.name" -}}
nodepool-operator
{{- end -}}

{{- define "priority-class.name" -}}
nodepool-critical
{{- end -}}

{{- define "priority-class.value" -}}
999999
{{- end -}}
