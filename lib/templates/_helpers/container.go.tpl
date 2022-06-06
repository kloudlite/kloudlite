{{- define "TemplateContainer" }}
{{- range $k, $v := . }}
{{- with $v}}
  - name: {{.Name}}
    image: {{.Image}}
    imagePullPolicy: {{.ImagePullPolicy | default "Always"}}
    {{- if .Env }}
    env:
    {{- include "TemplateEnv" .Env | indent 4}}
    {{- end }}
    resources:
    {{- if and .ResourceCpu.Min .ResourceMemory.Min }}
      requests:
      cpu: {{ .ResourceCpu.Min }}m
      memory: {{ .ResourceMemory.Min }}Mi
    {{- end }}
    {{- if and .ResourceCpu.Max .ResourceMemory.Max }}
      limits:
      cpu: {{ .ResourceCpu.Max }}m
      memory: {{ .ResourceMemory.Max }}Mi
    {{- end }}
{{- end }}
{{- end }}
{{- end }}
