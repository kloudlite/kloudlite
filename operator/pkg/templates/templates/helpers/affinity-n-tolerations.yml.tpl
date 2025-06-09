{{- define "NodeAffinity" -}}
nodeAffinity:
  preferredDuringSchedulingIgnoredDuringExecution:
    {{- $nWeight := 30 -}}
    {{- range $weight := Iterate $nWeight }}
    - weight: {{ sub $nWeight $weight }}
      preference:
        matchExpressions:
          - key: kloudlite.io/node-index
            operator: In
            values:
              - {{$weight | squote}}
    {{- end }}
{{- end }}

{{- define "RegionToleration" -}}
{{- $region := get . "region" -}}
- effect: NoExecute
  key: kloudlite.io/region
  operator: Equal
  value: {{$region}}
{{- end }}

{{- define "RegionNodeSelector" -}}
{{- $region := get . "region" -}}
kloudlite.io/region: {{$region}}
{{- end }}
