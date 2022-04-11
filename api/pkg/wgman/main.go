package wgman

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"text/template"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"kloudlite.io/pkg/rexec"
)

type wgManager struct {
	isServer     bool
	configPath   string
	remoteClient rexec.Rclient
}

func NewSshWgManager(configPath, hostIp, user, access_path string, isServer bool) *wgManager {
	return &wgManager{
		configPath:   configPath,
		remoteClient: rexec.NewSshRclient(hostIp, user, access_path),
		isServer:     isServer,
	}
}

func NewKubeWgManager(configPath, kubeconfigPath, namespace, name string, isServer bool) *wgManager {
	return &wgManager{
		configPath:   configPath,
		remoteClient: rexec.NewK8sRclient(kubeconfigPath, namespace, name),
		isServer:     isServer,
	}
}

type Peer struct {
	PublicKey  string  `json:"public_key"`
	Endpoint   *string `json:"endpoint,omitempty"`
	AllowedIps string  `json:"allowed_ips"`
}

type Config struct {
	PublicKey    string          `json:"public_key"`
	PrivateKey   string          `json:"private_key"`
	PublicIp     string          `json:"public_ip"`
	Peers        map[string]Peer `json:"peers"`
	NetInterface string          `json:"net_interface"`
}

func (c *Config) writeConfig(w wgManager) error {
	var config string
	var err error
	remoteClient := w.remoteClient
	configPath := w.configPath
	if w.isServer {
		config, err = c.getWgConfig()
	} else {
		config, err = c.getWgClientConfig()
	}
	// fmt.Println(config)
	if err != nil {
		return err
	}

	remoteClient.Run("wg-quick", "down", "wg0").Run()

	fmt.Println("Writing config to", configPath)
	// if err != nil {
	// 	return err
	// }

	remoteClient.Run("rm", configPath).Run()

	fmt.Print("Writing config to", configPath)
	if err != nil {
		panic(err)
		return err
	}

	err = remoteClient.WriteFile(configPath, []byte(config))
	fmt.Println("Writing config to", configPath, config)
	if err != nil {
		panic(err)
		return err
	}

	err = remoteClient.Run("wg-quick", "up", "wg0").Run()
	return err
}

func (c *Config) getWgConfig() (string, error) {
	f := `
[Interface]
Address ={{.PublicIp}}
SaveConfig = true
ListenPort = 31820
PrivateKey = {{.PrivateKey}}
PostUp = iptables -A FORWARD -i wg0 -j ACCEPT; iptables -t nat -A POSTROUTING -o {{ .NetInterface }} -j MASQUERADE; ip6tables -A FORWARD -i wg0 -j ACCEPT; ip6tables -t nat -A POSTROUTING -o {{ .NetInterface }} -j MASQUERADE
PostDown = iptables -D FORWARD -i wg0 -j ACCEPT; iptables -t nat -D POSTROUTING -o {{ .NetInterface }} -j MASQUERADE; ip6tables -D FORWARD -i wg0 -j ACCEPT; ip6tables -t nat -D POSTROUTING -o {{ .NetInterface }} -j MASQUERADE

{{- range $key, $value := .Peers }}

[Peer]
PublicKey = {{ $value.PublicKey }}
AllowedIPs = {{ $value.AllowedIps }}
	{{- if $value.Endpoint }}
Endpoint = {{ $value.Endpoint }}
	{{- end }}
{{- end}}
`
	parse, err := template.New("wg").Parse(f)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	parse.Execute(&buf, c)
	return buf.String(), err

}

func (c *Config) getWgClientConfig() (string, error) {
	f := `
[Interface]
Address ={{.PublicIp}}
PrivateKey = {{.PrivateKey}}

{{- range $key, $value := .Peers }}

[Peer]
PublicKey = {{ $value.PublicKey }}
AllowedIPs = {{ $value.AllowedIps }}
	{{- if $value.Endpoint }}
Endpoint = {{ $value.Endpoint }}
	{{- end }}
{{- end}}
`
	parse, err := template.New("wg").Parse(f)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	parse.Execute(&buf, c)
	return buf.String(), err

}

func (wg *wgManager) IsSetupDone() bool {
	_, err := wg.remoteClient.Readfile("/root/config.json")
	if err != nil {
		return false
	}
	return true
}

func (w *wgManager) Init(ip string) (*string, error) {
	fmt.Println("Initializing wireguard")

	key, err := wgtypes.GenerateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate public, private keys: %v", err)
	}

	c := Config{
		PublicKey:    key.PublicKey().String(),
		PrivateKey:   key.String(),
		PublicIp:     ip,
		Peers:        map[string]Peer{},
		NetInterface: "eth0",
	}

	var marshal []byte
	if marshal, err = json.Marshal(c); err != nil {
		fmt.Println("failed to marshal config: %v", err)
		return nil, fmt.Errorf("failed to masrshal config: %v ", err)
	}

	if err := ioutil.WriteFile("config.json", marshal, 0644); err != nil {
		return nil, fmt.Errorf("failed to write config: %v ", err)
	}

	if err := c.writeConfig(*w); err != nil {
		return nil, fmt.Errorf("unable to write config file: %v", err)
	}

	publicKey := key.PublicKey().String()

	fmt.Println("Wireguard initialized", publicKey)
	return &publicKey, nil

}

func (w *wgManager) GetNodeIp() (string, error) {
	out, err := w.remoteClient.Readfile("wg-ip")
	if err != nil {
		return "", fmt.Errorf("failed to get node ip: %v", err)
	}
	return string(out), nil
}

func (w1 *wgManager) connect(w2 wgManager) error {
	// c1, e := w1.getConfig()
	// if e != nil {
	// 	return e
	// }

	// c2, e := w2.getConfig()
	// if e != nil {
	// 	return e
	// }

	// w1.AddPeer(c2.PublicKey, "", nil, c2.NetInterface)

	// w2, e := w2.getConfig()

	return nil
}

func (w *wgManager) AddPeer(publicKey string, allowedIps string, endpoint *string, configPath string) error {
	var c Config
	var configsRaw []byte
	var err error
	if configsRaw, err = w.remoteClient.Readfile("config.json"); err != nil {
		return fmt.Errorf("unable to parse config error: %v", err)
	}
	if err := json.Unmarshal(configsRaw, &c); err != nil {
		return fmt.Errorf("unable to parse config error: %v", err)
	}
	c.Peers[publicKey] = Peer{
		PublicKey:  publicKey,
		Endpoint:   endpoint,
		AllowedIps: allowedIps,
	}
	return c.writeConfig(*w)
}

func (w *wgManager) getConfig() (Config, error) {
	var c Config
	var configsRaw []byte
	var err error
	if configsRaw, err = w.remoteClient.Readfile("config.json"); err != nil {
		return c, fmt.Errorf("unable to parse config error: %v", err)
	}
	if err := json.Unmarshal(configsRaw, &c); err != nil {
		return c, fmt.Errorf("unable to parse config error: %v", err)
	}
	return c, nil
}

func (w *wgManager) DeletePeer(publicKey string, configPath string) error {
	var c Config
	var configsRaw []byte
	var err error
	if configsRaw, err = w.remoteClient.Readfile("config.json"); err != nil {
		return fmt.Errorf("unable to parse config error: %v", err)
	}
	if err := json.Unmarshal(configsRaw, &c); err != nil {
		return fmt.Errorf("unable to parse config error: %v", err)
	}
	delete(c.Peers, publicKey)
	return c.writeConfig(*w)
}

// func main() {
// 	var ip, peersBase64, command, wgConfigPath string
// 	flag.StringVar(&ip, "ip", "", "public ip")
// 	flag.StringVar(&command, "command", "", "command")
// 	flag.StringVar(&peersBase64, "peers", "", "peers json in Base64")
// 	flag.StringVar(&wgConfigPath, "configpath", "/etc/wireguard/wg0.conf", "wg config path")

// 	flag.Parse()

// 	switch command {
// 	case "init":
// 		key, err := wgtypes.GenerateKey()
// 		if err != nil {
// 			panic(fmt.Errorf("failed to generate public, private keys: %v", err))
// 		}
// 		c := Config{
// 			PublicKey:    key.PublicKey().String(),
// 			PrivateKey:   key.String(),
// 			PublicIp:     ip,
// 			Peers:        map[string]Peer{},
// 			NetInterface: "eth0",
// 		}
// 		marshal, err := json.Marshal(c)
// 		if err != nil {
// 			panic(fmt.Errorf("failed to marshal config: %v", err))
// 		}
// 		err = exec.Command("rm", "./wg0.conf").Run()
// 		err = ioutil.WriteFile("config.json", marshal, 0644)
// 		if err != nil {
// 			panic(fmt.Errorf("unable to write config file: %v", err))
// 		}
// 		out, err := json.Marshal(map[string]string{
// 			"public_key": key.PublicKey().String(),
// 		})
// 		if err != nil {
// 			panic(fmt.Errorf("unable to generate output: %v", err))
// 		}
// 		c.writeConfig(wgConfigPath)
// 		fmt.Println(string(out))
// 		break
// 	case "peers":

// 		// all, _ := io.ReadAll(os.Stdin)
// 		// os.Stdin.Close()
// 		all, _ := base64.StdEncoding.DecodeString(peersBase64)
// 		peers := make([]Peer, 0)
// 		err := json.Unmarshal(all, &peers)
// 		if err != nil {
// 			fmt.Println(fmt.Errorf("unable to parse peers error: %v", err))
// 		}
// 		var c Config
// 		configsRaw, err := ioutil.ReadFile("config.json")
// 		err = json.Unmarshal(configsRaw, &c)
// 		if err != nil {
// 			fmt.Println(fmt.Errorf("unable to parse config error: %v", err))
// 		}

// 		for k := range c.Peers {
// 			delete(c.Peers, k)
// 		}

// 		for _, p := range peers {
// 			c.Peers[p.PublicKey] = p
// 		}

// 		marshal, err := json.Marshal(c)
// 		if err != nil {
// 			panic(fmt.Errorf("failed to marshal config: %v", err))
// 		}

// 		err = c.writeConfig(wgConfigPath)

// 		if err != nil {
// 			panic(fmt.Errorf("unable to update wireguard: %v", err))
// 		}

// 		err = ioutil.WriteFile("config.json", marshal, 0644)

// 		if err != nil {
// 			panic(fmt.Errorf("unable to write config file: %v", err))
// 		}

// 		break
// 	}
// }
