package wgutils

import (
	"github.com/kloudlite/operator/pkg/templates"
)

type PublicPeer struct {
	PublicKey  string
	AllowedIPs []string
	Endpoint   string
	IPAddr     string
}

type PrivatePeer struct {
	PublicKey  string
	AllowedIPs []string
}

type WgConfigParams struct {
	IPAddr     string
	PrivateKey string
	DNS        string

	PostUp   []string
	PostDown []string

	PublicPeers  []PublicPeer
	PrivatePeers []PrivatePeer
}

func GenerateWireguardConfig(wgParams WgConfigParams) (string, error) {
	t := templates.NewTextTemplate("wg-config")
	b, err := t.ParseBytes([]byte(`
[Interface]
Address = {{.IPAddr}}/32
PrivateKey = {{.PrivateKey}}
DNS = {{.DNS}}

{{- range .PostUp -}}
PostUp = {{.}}
{{- end -}}

{{- range .PostDown -}}
PostDown = {{.}}
{{- end -}}

{{- range .PublicPeers}}
{{- with .}}
[Peer]
PublicKey = {{.PublicKey}}
AllowedIPs = {{.AllowedIPs | join ", " }}, {{.IPAddr}}/32
Endpoint = {{.Endpoint}}
{{- end }}
{{- end }}

{{- range .PrivatePeers}}
{{- with .}}
[Peer]
PublicKey = {{.PublicKey}}
AllowedIPs = {{.AllowedIPs | join ", " }}
{{- end }}
{{- end }}
`), wgParams)
	if err != nil {
		return "", err
	}

	return string(b), nil
}
