[Interface]
Address = {{ .IpAddress}}
ListenPort = 51820
PrivateKey = {{ .PrivateKey }}

# Allow DNS traffic to pass through the
PostUp = iptables -A FORWARD -i eth0 -p udp --dport 53 -j ACCEPT
PostUp = iptables -t nat -A PREROUTING -i %i -p udp --dport 53 -j DNAT --to {{ .DnsServer }}:53

# Wireguard
PostUp = iptables -A FORWARD -i eth0 -j ACCEPT; iptables -t nat -A POSTROUTING -o %i -j MASQUERADE
PostUp = iptables -A FORWARD -i %i -j ACCEPT; iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE


# Dns
PostDown = iptables -D FORWARD -i eth0 -p udp --dport 53 -j ACCEPT
PostDown = iptables -t nat -D PREROUTING -i %i -p udp --dport 53 -j DNAT --to {{ .DnsServer }}:53

# Wireguard
PostDown = iptables -D FORWARD -i %i -j ACCEPT; iptables -t nat -D POSTROUTING -o eth0 -j MASQUERADE
PostDown = iptables -D FORWARD -i eth0 -j ACCEPT; iptables -t nat -D POSTROUTING -o %i -j MASQUERADE

# forward to services

{{- range $virtual, $actual := .IpForwardingMap }}
PostUp = iptables -t nat -A PREROUTING -i %i -d {{ $virtual }} -j DNAT --to-destination {{ $actual }}
PostDown = iptables -t nat -D PREROUTING -i %i -d {{ $virtual }} -j DNAT --to-destination {{ $actual }}
{{- end }}

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
