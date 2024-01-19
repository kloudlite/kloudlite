package wgc

import (
	"fmt"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/table"
	"github.com/kloudlite/kl/pkg/ui/text"
	"math"
	"net"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

const (
	EnvWgCommand  = "WG_COMMAND"
	EnvWgHideKeys = "WG_HIDE_KEYS"
)

type WgShowOptions struct {
	Interface string
	Option    string
	ShowKeys  bool
}

func getOptions() *WgShowOptions {
	opts := WgShowOptions{}
	opts.ShowKeys = os.Getenv(EnvWgHideKeys) == "never"
	opts.Interface = "all"
	return &opts
}

func Show(opts *WgShowOptions) ([]string, error) {
	res := []string{}
	// opts := getOptions()
	if opts == nil {
		opts = getOptions()
	}

	client, err := wgctrl.New()
	if err != nil {
		return nil, err
	}

	// checkError(err)
	switch opts.Interface {
	case "interfaces":
		devices, err := client.Devices()
		if err != nil {
			return nil, err
		}
		for i := 0; i < len(devices); i++ {
			res = append(res, devices[i].Name)
		}
	case "all":
		devices, err := client.Devices()
		if err != nil {
			return nil, err
		}
		for _, dev := range devices {
			err := showDevice(*dev, opts)
			if err != nil {
				return nil, err
			}
		}
	default:
		dev, err := client.Device(opts.Interface)
		if err != nil {
			return nil, err
		}

		err = showDevice(*dev, opts)
		if err != nil {
			return nil, err
		}
	}

	return res, client.Close()
}

func showDevice(dev wgtypes.Device, opts *WgShowOptions) error {
	if opts.Option == "" {
		showKeys := opts.ShowKeys
		fn.Println("")
		fn.Println(text.Bold(text.Green("Interface:")), text.Red(fmt.Sprintf("%s (%s)", dev.Name, dev.Type.String())))
		table.KVOutput("  public key:", text.Colored(dev.PublicKey.String(), 4), true)
		table.KVOutput("  private key:", formatKey(dev.PrivateKey, showKeys), true)
		table.KVOutput("  listening port:", text.Colored(dev.ListenPort, 2), true)
		fn.Println("")

		for _, peer := range dev.Peers {
			err := showPeers(peer, showKeys)
			if err != nil {
				return err
			}
		}
	} else {
		deviceName := ""
		if opts.Interface == "all" {
			deviceName = dev.Name + "\t"
		}
		switch opts.Option {
		case "public-key":
			fn.Printf("%s%s\n", deviceName, dev.PublicKey.String())
		case "private-key":
			fn.Printf("%s%s\n", deviceName, dev.PrivateKey.String())
		case "listen-port":
			fn.Printf("%s%d\n", deviceName, dev.ListenPort)
		case "fwmark":
			fn.Printf("%s%d\n", deviceName, dev.FirewallMark)
		case "peers":
			for _, peer := range dev.Peers {
				fn.Printf("%s%s\n", deviceName, peer.PublicKey.String())
			}
		case "preshared-keys":
			for _, peer := range dev.Peers {
				fn.Printf("%s%s\t%s\n", deviceName, peer.PublicKey.String(), formatPSK(peer.PresharedKey, "(none)"))
			}
		case "endpoints":
			for _, peer := range dev.Peers {
				fn.Printf("%s%s\t%s\n", deviceName, peer.PublicKey.String(), formatEndpoint(peer.Endpoint))
			}
		case "allowed-ips":
			for _, peer := range dev.Peers {
				fn.Printf("%s%s\t%s\n", deviceName, peer.PublicKey.String(), joinIPs(peer.AllowedIPs))
			}
		case "latest-handshakes":
			for _, peer := range dev.Peers {
				fn.Printf("%s%s\t%d\n", deviceName, peer.PublicKey.String(), peer.LastHandshakeTime.Unix())
			}
		case "transfer":
			for _, peer := range dev.Peers {
				fn.Printf("%s%s\t%d\t%d\n", deviceName, peer.PublicKey.String(), peer.ReceiveBytes, peer.TransmitBytes)
			}
		case "persistent-keepalive":
			for _, peer := range dev.Peers {
				fn.Printf("%s%s\t%s\n", deviceName, peer.PublicKey.String(), zeroToOff(strconv.FormatFloat(peer.PersistentKeepaliveInterval.Seconds(), 'g', 0, 64)))
			}
		case "dump":
			fn.Printf("%s%s\t%s\t%d\t%s\n", deviceName, dev.PrivateKey.String(), dev.PublicKey.String(), dev.ListenPort, zeroToOff(strconv.FormatInt(int64(dev.FirewallMark), 10)))
			for _, peer := range dev.Peers {
				fn.Printf("%s%s\t%s\t%s\t%s\t%d\t%d\t%d\t%s\n",
					deviceName,
					peer.PublicKey.String(),
					formatPSK(peer.PresharedKey, "(none)"),
					formatEndpoint(peer.Endpoint),
					joinIPs(peer.AllowedIPs),
					peer.LastHandshakeTime.Unix(),
					peer.ReceiveBytes,
					peer.TransmitBytes,
					zeroToOff(strconv.FormatFloat(peer.PersistentKeepaliveInterval.Seconds(), 'g', 0, 64)))
			}
		}
	}
	return nil
}

func showPeers(peer wgtypes.Peer, showKeys bool) error {
	/*
			 keep alive interval = {{ .KeepAliveInterval }}s
		  protocol version = {{ .ProtocolVersion }}
	*/

	tmpl := fmt.Sprintf(`%s %s 
  %s {{ .Endpoint }}
  %s {{ .AllowedIPs }}
  {{- if .PresharedKey}}
  %s {{ .PresharedKey }}
  {{- end}} 
  %s {{ .LastHandshakeTime }}
  %s %s
`, text.Bold(text.Red("peers: ")), text.Blue("{{ .PublicKey }}"),
		text.Bold("endpoint = "),
		text.Bold("allowed ips ="),
		text.Bold("preshared key ="),
		text.Bold("last handshake ="),
		text.Bold("transfer:"),
		fmt.Sprintf("%s received, %s sent",
			text.Green("{{ .ReceiveBytes }}"),
			text.Green("{{ .TransmitBytes }}")),
	)
	type tmplContent struct {
		PublicKey         string
		PresharedKey      string
		Endpoint          string
		KeepAliveInterval float64
		LastHandshakeTime string
		ReceiveBytes      string
		TransmitBytes     string
		AllowedIPs        string
		ProtocolVersion   int
	}

	t := template.Must(template.New("peer_tmpl").Parse(tmpl))
	c := tmplContent{
		PublicKey:         peer.PublicKey.String(),
		PresharedKey:      formatPSK(peer.PresharedKey, ""),
		Endpoint:          formatEndpoint(peer.Endpoint),
		KeepAliveInterval: peer.PersistentKeepaliveInterval.Seconds(),
		LastHandshakeTime: fromNow(peer.LastHandshakeTime),
		ReceiveBytes:      ByteCountIEC(peer.ReceiveBytes),
		TransmitBytes:     ByteCountIEC(peer.TransmitBytes),
		AllowedIPs:        joinIPs(peer.AllowedIPs),
		ProtocolVersion:   peer.ProtocolVersion,
	}

	err := t.Execute(os.Stdout, c)
	if err != nil {
		return err
	}
	return nil
}

func formatEndpoint(endpoint *net.UDPAddr) string {
	ip := endpoint.String()
	if ip == "<nil>" {
		ip = "(none)"
	}
	return ip
}

func formatKey(key wgtypes.Key, showKeys bool) string {
	k := "(hidden)"
	if showKeys {
		k = key.String()
	}
	return k
}

func formatPSK(key wgtypes.Key, none string) string {
	psk := key.String()
	if psk == "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=" {
		return none
	}
	return psk
}

func joinIPs(ips []net.IPNet) string {
	ipStrings := make([]string, 0, len(ips))
	for _, v := range ips {
		ipStrings = append(ipStrings, v.String())
	}
	return strings.Join(ipStrings, ", ")
}

func zeroToOff(value string) string {
	if value == "0" {
		return "off"
	}
	return value
}

func ByteCountIEC(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB",
		float64(b)/float64(div), "KMGTPE"[exp])
}

func fromNow(t2 time.Time) string {
	t1 := time.Now()

	hs := t1.Sub(t2).Hours()

	hs, mf := math.Modf(hs)
	ms := mf * 60

	ms, sf := math.Modf(ms)
	ss := sf * 60

	if hs > 500000 {
		hs = 0
		ms = 0
		ss = 0
	}
	return fmt.Sprintf("%s hours %s minutes %s seconds ago",
		text.Bold(text.Red(int(hs))),
		text.Bold(text.Red(int(ms))),
		text.Bold(text.Red(int(ss))),
	)
}
