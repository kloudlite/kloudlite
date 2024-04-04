[Interface]
Address = {{ .IpAddress}}
PrivateKey = {{ .PrivateKey }}

{{- range $_, $peer := .Peers }}
{{ with $peer }}
[Peer]
PublicKey = {{ .PublicKey }}
{{- if .Endpoint }}
Endpoint = {{ .Endpoint }}
{{- end}}

{{- if .AllowedIPs }}
AllowedIPs = {{ range $i, $ip := .AllowedIPs }}
{{- if $i}}, {{- end}}
{{- $ip }}
{{- end}}
{{- end}}

{{- end }}
{{- end }}
