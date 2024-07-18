package wgutils

import (
	"strings"

	"github.com/kloudlite/operator/pkg/templates"
)

type PublicPeer struct {
	// displayname is used in comment for the specified peer
	DisplayName string

	PublicKey  string
	AllowedIPs []string
	Endpoint   string
}

type PrivatePeer struct {
	// displayname is used in comment for the specified peer
	DisplayName string

	PublicKey  string
	AllowedIPs []string
}

type WgConfigParams struct {
	IPAddr     string
	PrivateKey string
	ListenPort uint16

	DNS string

	PostUp   []string
	PostDown []string

	PublicPeers  []PublicPeer
	PrivatePeers []PrivatePeer
}

func GenerateWireguardConfig(wgParams WgConfigParams) (string, error) {
	t := templates.NewTextTemplate("wg-config")
	b, err := t.ParseBytes([]byte(strings.TrimSpace(`
[Interface]
Address = {{.IPAddr}}/32
PrivateKey = {{.PrivateKey}}
{{- if .ListenPort }}
ListenPort = {{.ListenPort}}
{{- end }}

{{- /* {{- if .DNS }} */}}
{{- /* DNS = {{.DNS}} */}}
{{- /* {{- end }} */}}

{{- range .PostUp -}}
PostUp = {{.}}
{{- end -}}

{{- range .PostDown -}}
PostDown = {{.}}
{{- end -}}

{{- range .PublicPeers}}
{{- with .}}
[Peer]
{{- if .DisplayName }}
# {{.DisplayName}}
{{- end }} 
PublicKey = {{.PublicKey}}
AllowedIPs = {{.AllowedIPs | join ", " }}
{{- if .Endpoint }}
Endpoint = {{.Endpoint}}
{{- end }}
PersistentKeepalive = 25
{{- end }}
{{- end }}

{{- range .PrivatePeers}}
{{- with .}}
[Peer]
{{- if .DisplayName }}
# {{.DisplayName}}
{{- end }} 
PublicKey = {{.PublicKey}}
AllowedIPs = {{.AllowedIPs | join ", " }}
PersistentKeepalive = 25
{{- end }}
{{- end }}
`)), wgParams)
	if err != nil {
		return "", err
	}

	return string(b), nil
}
