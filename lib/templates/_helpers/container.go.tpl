{{- define "TemplateContainer" }}
{{- $containers := get . "containers"}}
{{- $volumeMounts := get . "volumeMounts"}}
{{- range $idx, $v := $containers }}
{{- with $v }}
  - name: {{.Name}}
    image: {{.Image}}
    imagePullPolicy: {{.ImagePullPolicy | default "IfNotPresent"}}

  {{- if .Env }}
    env:
    {{- include "TemplateEnv" .Env | indent 4}}
  {{- end }}

  {{- if .EnvFrom }}
    envFrom:
    {{- range .EnvFrom }}
      {{- if .Config }}
      - configMapRef:
          name: {{.Config}}
      {{- end }}
      {{- if .Secret }}
      - secretRef:
          name: {{.Secret}}
      {{- end }}
    {{- end}}
  {{- end}}

  {{- if or .ResourceCpu .ResourceMemory }}
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

  {{ $vMounts := index $volumeMounts $idx }}
  {{- if $vMounts }}
    volumeMounts: {{- $vMounts | toPrettyJson | indent 4 }}
{{/*    {{- range $v := $vMounts }}*/}}
{{/*      - name: {{$v.Name}}*/}}
{{/*        mountPath: {{$v.MountPath}}*/}}
{{/*    {{- end }}*/}}
  {{- end}}
{{- end }}
{{- end }}
{{- end }}
