{{- define "service-account-name" -}}
{{.Release.Name}}-{{.Values.serviceAccount.nameSuffix}}
{{- end -}}

