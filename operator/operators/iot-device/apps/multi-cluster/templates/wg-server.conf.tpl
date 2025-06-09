[Interface]
Address = {{ .IpAddress}}
ListenPort = 51820
PrivateKey = {{ .PrivateKey }}

PostUp = iptables -A FORWARD -i eth0 -j ACCEPT; iptables -t nat -A POSTROUTING -o %i -j MASQUERADE
PostUp = iptables -A FORWARD -i %i -j ACCEPT; iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE

PostDown = iptables -D FORWARD -i %i -j ACCEPT; iptables -t nat -D POSTROUTING -o eth0 -j MASQUERADE
PostDown = iptables -D FORWARD -i eth0 -j ACCEPT; iptables -t nat -D POSTROUTING -o %i -j MASQUERADE

{{- range $_, $peer := .Peers }}
{{ with $peer }}
[Peer]
PublicKey = {{ .PublicKey }}

{{- if .AllowedIPs }}
AllowedIPs = {{ range $i, $ip := .AllowedIPs }}
{{- if $i}}, {{- end}}
{{- $ip }}
{{- end}}
{{- end}}

{{- if .Endpoint }}
Endpoint = {{ .Endpoint }}
{{- end}}

{{- end }}
{{- end }}

{{- range $_, $peer := .InternalPeers }}
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
