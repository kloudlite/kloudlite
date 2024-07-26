{{- /* platformoperator */}}
{{- if not .Values.operators.platformOperator.configuration.IACStateStore.s3BucketName}}}
{{ fail ".operators.platformOperator.configuration.cluster.IACStateStore.s3BucketName is required" }}
{{- end }}
