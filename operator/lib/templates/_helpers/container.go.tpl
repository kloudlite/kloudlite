{{- define "TemplateContainer" }}
{{- $containers := get . "containers"}}
{{- $volumeMounts := get . "volumeMounts"}}

{{- range $idx, $v := $containers }}
{{- with $v }}
- name: {{.Name}}
  image: {{.Image}}
  imagePullPolicy: {{.ImagePullPolicy | default "IfNotPresent"}}
  {{- if .Command }}
  command: {{.Command | toYAML | nindent 4 }}
  {{- end}}

  {{- if .Args }}
  args: {{.Args | toYAML | nindent 4}}
  {{- end}}

  {{- if .Env }}
  env: {{- include "TemplateEnv" .Env | indent 6 }}
  {{- end }}

  {{- if .EnvFrom }}
  envFrom:
  {{- range .EnvFrom }}
    {{- if eq .Type "config" }}
    - configMapRef:
        name: {{.RefName}}
    {{- end }}
    {{- if eq .Type "secret" }}
    - secretRef:
        name: {{.RefName}}
    {{- end }}
  {{- end}}
  {{- end}}

  {{- if or .ResourceCpu .ResourceMemory }}
  resources:
  {{- if and .ResourceCpu.Min .ResourceMemory.Min }}
    requests:
      cpu: {{ .ResourceCpu.Max }}
      memory: {{ .ResourceMemory.Max }}
  {{- end }}
  {{- if and .ResourceCpu.Max .ResourceMemory.Max }}
    limits:
      cpu: {{ .ResourceCpu.Max }}
      memory: {{ .ResourceMemory.Max }}
  {{- end }}
  {{- end }}

  {{- if $volumeMounts }}
  {{- $vMounts := index $volumeMounts $idx }}
  {{- if $vMounts }}
  volumeMounts: {{- $vMounts | toYAML | nindent 4 }}
  {{- end}}
  {{- end }}

  {{- if .LivenessProbe }}
  {{- with .LivenessProbe}}
  livenessProbe:
    failureThreshold: {{.FailureThreshold | default 3}}
    initialDelaySeconds: {{.InitialDelay | default 2}}
    periodSeconds: {{.Interval | default 10 }}

    {{- if eq .Type "shell"}}
    exec:
      command: {{ .Shell | toYAML | nindent 8 }}
    {{- end }}

    {{- if eq .Type "httpGet"}}
    httpGet: {{.HttpGet | toYAML | nindent 6}}
    {{- end }}

    {{- if eq .Type "httpHeaders"}}
    tcpProbe: {{.Tcp | toYAML | nindent 6}}
    {{- end}}
  {{- end }}
  {{- end}}

  {{- if .ReadinessProbe }}
  {{- with .ReadinessProbe}}
  readinessProbe:
    failureThreshold: {{.FailureThreshold | default 3}}
    initialDelaySeconds: {{.InitialDelay | default 2}}
    periodSeconds: {{.Interval | default 10 }}

    {{- if eq .Type "shell"}}
    exec:
      command: {{ .Shell | toYAML | nindent 8 }}
    {{- end }}

    {{- if eq .Type "httpGet"}}
    httpGet: {{.HttpGet | toYAML | nindent 6}}
    {{- end }}

    {{- if eq .Type "httpHeaders"}}
    tcpProbe: {{.Tcp | toYAML | nindent 6}}
    {{- end}}
  {{- end }}
  {{- end}}

{{- end }}
{{- end }}
{{- end }}
