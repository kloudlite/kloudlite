{{- define "TemplateContainer" }}
{{- $containers := get . "containers"}}
{{- $volumeMounts := get . "volumeMounts"}}

{{- range $idx, $v := $containers }}
{{- with $v }}
  - name: {{.Name}}
    image: {{.Image}}
    imagePullPolicy: {{.ImagePullPolicy | default "IfNotPresent"}}
    {{- if .Command }}
    command: {{.Command | toPrettyJson | indent 2 }}
    {{- end}}

    {{- if .Args }}
    args: {{.Args | toPrettyJson | indent 2}}
    {{- end}}

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
  {{- end}}

    readinessProbe:
      periodSeconds: 0

    livenessProbe:
      initialDelaySeconds: 3
      periodSeconds: 1
      httpGet:
        path: /
        port: 8080
        scheme: HTTP

{{- end }}
{{- end }}
{{- end }}
