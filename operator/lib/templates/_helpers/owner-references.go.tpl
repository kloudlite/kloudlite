{{- define "TemplateOwnerRefs" }}
{{- $ownerRefs := . }}
{{- if $ownerRefs }}
{{$ownerRefs | toJson | toYAML}}
{{- end }}
{{- end }}
