package wgc

import (
	"bytes"
	"encoding"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/kloudlite/kl/lib/common"
	"github.com/kloudlite/kl/lib/common/ui/color"
	"github.com/lixiangzhong/dnsutil"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// Config represents full wg-quick like config structure
type Config struct {
	wgtypes.Config

	// Address list of IP (v4 or v6) addresses (optionally with CIDR masks) to be assigned to the interface. May be specified multiple times.
	Address []net.IPNet

	// list of IP (v4 or v6) addresses to be set as the interface’s DNS servers. May be specified multiple times. Upon bringing the interface up, this runs ‘resolvconf -a tun.INTERFACE -m 0 -x‘ and upon bringing it down, this runs ‘resolvconf -d tun.INTERFACE‘. If these particular invocations of resolvconf(8) are undesirable, the PostUp and PostDown keys below may be used instead.
	DNS []net.IP

	// MTU is automatically determined from the endpoint addresses or the system default route, which is usually a sane choice. However, to manually specify an MTU to override this automatic discovery, this value may be specified explicitly.
	MTU int

	// Table — Controls the routing table to which routes are added.
	Table int

	// PreUp, PostUp, PreDown, PostDown — script snippets which will be executed by bash(1) before/after setting up/tearing down the interface, most commonly used to configure custom DNS options or firewall rules. The special string ‘%i’ is expanded to INTERFACE. Each one may be specified multiple times, in which case the commands are executed in order.
	PreUp    string
	PostUp   string
	PreDown  string
	PostDown string

	// RouteProtocol to set on the route. See linux/rtnetlink.h  Use value > 4 or default 0
	RouteProtocol int

	// RouteMetric sets this metric on all managed routes. Lower number means pick this one
	RouteMetric int

	// Address label to set on the link
	AddressLabel string

	// SaveConfig — if set to ‘true’, the configuration is saved from the current state of the interface upon shutdown.
	// Currently unsupported
	SaveConfig bool
}

var _ encoding.TextMarshaler = (*Config)(nil)
var _ encoding.TextUnmarshaler = (*Config)(nil)

func (cfg *Config) String() string {
	b, err := cfg.MarshalText()
	if err != nil {
		panic(err)
	}
	return string(b)
}

func serializeKey(key *wgtypes.Key) string {
	return base64.StdEncoding.EncodeToString(key[:])
}

func toSeconds(duration time.Duration) int {
	return int(duration / time.Second)
}

var funcMap = template.FuncMap(map[string]interface{}{
	"wgKey":     serializeKey,
	"toSeconds": toSeconds,
})

var cfgTemplate = template.Must(
	template.
		New("wg-cfg").
		Funcs(funcMap).
		Parse(wgtypeTemplateSpec))

func (cfg *Config) MarshalText() (text []byte, err error) {
	buff := &bytes.Buffer{}
	if err := cfgTemplate.Execute(buff, cfg); err != nil {
		return nil, err
	}
	return buff.Bytes(), nil
}

const wgtypeTemplateSpec = `[Interface]
{{- range .Address }}
Address = {{ . }}
{{- end }}
{{- range .DNS }}
DNS = {{ . }}
{{- end }}
PrivateKey = {{ .PrivateKey | wgKey }}
{{- if .ListenPort }}{{ "\n" }}ListenPort = {{ .ListenPort }}{{ end }}
{{- if .MTU }}{{ "\n" }}MTU = {{ .MTU }}{{ end }}
{{- if .Table }}{{ "\n" }}Table = {{ .Table }}{{ end }}
{{- if .PreUp }}{{ "\n" }}PreUp = {{ .PreUp }}{{ end }}
{{- if .PostUp }}{{ "\n" }}PostUp = {{ .PostUp }}{{ end }}
{{- if .PreDown }}{{ "\n" }}PreDown = {{ .PreDown }}{{ end }}
{{- if .PostDown }}{{ "\n" }}PostDown = {{ .PostDown }}{{ end }}
{{- if .SaveConfig }}{{ "\n" }}SaveConfig = {{ .SaveConfig }}{{ end }}
{{- range .Peers }}
{{- "\n" }}
[Peer]
PublicKey = {{ .PublicKey | wgKey }}
AllowedIPs = {{ range $i, $el := .AllowedIPs }}{{if $i}}, {{ end }}{{ $el }}{{ end }}
{{- if .PresharedKey }}{{ "\n" }}PresharedKey = {{ .PresharedKey }}{{ end }}
{{- if .PersistentKeepaliveInterval }}{{ "\n" }}PersistentKeepalive = {{ .PersistentKeepaliveInterval | toSeconds }}{{ end }}
{{- if .Endpoint }}{{ "\n" }}Endpoint = {{ .Endpoint }}{{ end }}
{{- end }}
`

// ParseKey parses the base64 encoded wireguard private key
func ParseKey(key string) (wgtypes.Key, error) {
	var pkey wgtypes.Key
	pkeySlice, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return pkey, err
	}
	copy(pkey[:], pkeySlice[:])
	return pkey, nil
}

type parseState int

const (
	unknown parseState = iota
	inter              = iota
	peer               = iota
)

func (cfg *Config) UnmarshalText(text []byte) error {
	*cfg = Config{} // Zero out the config
	state := unknown
	var peerCfg *wgtypes.PeerConfig
	for no, line := range strings.Split(string(text), "\n") {
		ln := strings.TrimSpace(line)
		if len(ln) == 0 || ln[0] == '#' {
			continue
		}
		switch ln {
		case "[Interface]":
			state = inter
		case "[Peer]":
			state = peer
			cfg.Peers = append(cfg.Peers, wgtypes.PeerConfig{})
			peerCfg = &cfg.Peers[len(cfg.Peers)-1]
		default:
			parts := strings.Split(ln, "=")
			if len(parts) < 2 {
				return fmt.Errorf("cannot parse line %d, missing =", no)
			}
			lhs := strings.TrimSpace(parts[0])
			rhs := strings.TrimSpace(strings.Join(parts[1:], "="))

			switch state {
			case inter:
				if err := parseInterfaceLine(cfg, lhs, rhs); err != nil {
					return fmt.Errorf("[line %d]: %v", no+1, err)
				}
			case peer:
				if err := parsePeerLine(peerCfg, lhs, rhs); err != nil {
					return fmt.Errorf("[line %d]: %v", no+1, err)
				}
			default:
				return fmt.Errorf("[line %d] cannot parse, unknown state", no+1)
			}

		}
	}

	return nil
}
func parseInterfaceLine(cfg *Config, lhs string, rhs string) error {
	switch lhs {
	case "Address":
		for _, addr := range strings.Split(rhs, ",") {
			ip, cidr, err := net.ParseCIDR(strings.TrimSpace(addr))
			if err != nil {
				return err
			}
			cfg.Address = append(cfg.Address, net.IPNet{IP: ip, Mask: cidr.Mask})
		}
	case "DNS":
		for _, addr := range strings.Split(rhs, ",") {
			ip := net.ParseIP(strings.TrimSpace(addr))
			if ip == nil {
				return fmt.Errorf("cannot parse IP")
			}
			cfg.DNS = append(cfg.DNS, ip)
		}
	case "MTU":
		mtu, err := strconv.ParseInt(rhs, 10, 64)
		if err != nil {
			return err
		}
		cfg.MTU = int(mtu)
	case "Table":
		tbl, err := strconv.ParseInt(rhs, 10, 64)
		if err != nil {
			return err
		}
		cfg.Table = int(tbl)
	case "ListenPort":
		portI64, err := strconv.ParseInt(rhs, 10, 64)
		if err != nil {
			return err
		}
		port := int(portI64)
		cfg.ListenPort = &port
	case "PreUp":
		cfg.PreUp = rhs
	case "PostUp":
		cfg.PostUp = rhs
	case "PreDown":
		cfg.PreDown = rhs
	case "PostDown":
		cfg.PostDown = rhs
	case "SaveConfig":
		save, err := strconv.ParseBool(rhs)
		if err != nil {
			return err
		}
		cfg.SaveConfig = save
	case "PrivateKey":
		key, err := ParseKey(rhs)
		if err != nil {
			return fmt.Errorf("cannot decode key %v", err)
		}
		cfg.PrivateKey = &key
	default:
		return fmt.Errorf("unknown directive %s", lhs)
	}
	return nil
}

func parsePeerLine(peerCfg *wgtypes.PeerConfig, lhs string, rhs string) error {
	switch lhs {
	case "PublicKey":
		key, err := ParseKey(rhs)
		if err != nil {
			return fmt.Errorf("cannot decode key %v", err)
		}
		peerCfg.PublicKey = key
	case "PresharedKey":
		key, err := ParseKey(rhs)
		if err != nil {
			return fmt.Errorf("cannot decode key %v", err)
		}
		if peerCfg.PresharedKey != nil {
			return fmt.Errorf("preshared key already defined %v", err)
		}
		peerCfg.PresharedKey = &key
	case "AllowedIPs":
		for _, addr := range strings.Split(rhs, ",") {
			ip, cidr, err := net.ParseCIDR(strings.TrimSpace(addr))
			if err != nil {
				return fmt.Errorf("cannot parse %s: %v", addr, err)
			}
			peerCfg.AllowedIPs = append(peerCfg.AllowedIPs, net.IPNet{IP: ip, Mask: cidr.Mask})
		}
	case "Endpoint":

		udpAddr, err2 := func() (*net.UDPAddr, error) {
			var dig dnsutil.Dig
			dig.SetDNS("1.1.1.1") //or ns.xxx.com
			strings.Split(rhs, ":")
			a, err := dig.A(strings.Split(rhs, ":")[0]) // dig google.com @8.8.8.8
			if err != nil {
				return nil, err
			}
			if len(a) == 0 {
				common.PrintError(errors.New(color.Text("defering to net.ResolveUDPAddr", 209)))
				return net.ResolveUDPAddr("", rhs)
			}
			port, err := strconv.ParseInt(strings.Split(rhs, ":")[1], 10, 32)
			if err != nil {
				common.PrintError(errors.New(color.Text("defering to net.ResolveUDPAddr", 209)))
				return net.ResolveUDPAddr("", rhs)
			}
			return &net.UDPAddr{
				IP:   a[0].A.To4(),
				Port: int(port),
			}, nil
		}()
		if err2 != nil {
			return err2
		}

		peerCfg.Endpoint = udpAddr
	case "PersistentKeepalive":
		t, err := strconv.ParseInt(rhs, 10, 64)
		if err != nil {
			return err
		}
		dur := time.Duration(t * int64(time.Second))
		peerCfg.PersistentKeepaliveInterval = &dur
	default:
		return fmt.Errorf("unknown directive %s", lhs)
	}
	return nil
}
