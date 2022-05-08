package wgman

import (
	"bytes"
	"encoding/json"
	"fmt"
	"text/template"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"kloudlite.io/pkg/rexec"
)

type wgManager struct {
	isServer       bool
	configPath     string
	remoteClient   rexec.Rclient
	configDataPath string
}

func NewSshWgManager(configPath, hostIp, user, accessPath, configDataPath string, isServer bool) *wgManager {
	return &wgManager{
		configPath:     configPath,
		configDataPath: configDataPath,
		remoteClient:   rexec.NewSshRclient(hostIp, user, accessPath),
		isServer:       isServer,
	}
}

func NewKubeWgManager(configPath, kubeConfigPath, namespace, name, configDataPath string, isServer bool) *wgManager {
	return &wgManager{
		configPath:     configPath,
		configDataPath: configDataPath,
		remoteClient:   rexec.NewK8sRclient(kubeConfigPath, namespace, name),
		isServer:       isServer,
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
	RemotePeers  map[string]Peer `json:"remote_peers"`
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
		panic(err)
		return err
	}

	remoteClient.Run("wg-quick", "down", "wg0").Run()
	remoteClient.Run("rm", configPath).Run()
	if err != nil {
		panic(err)
		return err
	}
	err = remoteClient.WriteFile(configPath, []byte(config))
	if err != nil {
		panic(err)
		return err
	}

	err = remoteClient.Run("wg-quick", "up", "wg0").Run()

	if err != nil {
		panic(err)
	}
	return err
}

func (c *Config) getWgConfig() (string, error) {
	f := `
[Interface]
Address ={{.PublicIp}}
ListenPort = 51820
PrivateKey = {{.PrivateKey}}
PostUp = iptables -A FORWARD -i %i -j ACCEPT; iptables -A FORWARD -o %i -j ACCEPT; iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE
PostDown = iptables -D FORWARD -i %i -j ACCEPT; iptables -D FORWARD -o %i -j ACCEPT; iptables -t nat -D POSTROUTING -o eth0 -j MASQUERADE


{{- range $key, $value := .Peers }}

[Peer]
PublicKey = {{ $value.PublicKey }}
AllowedIPs = {{ $value.AllowedIps }}
	{{- if $value.Endpoint }}
Endpoint = {{ $value.Endpoint }}
	{{- end }}
{{- end}}

{{- range $key, $value := .RemotePeers }}

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
PrivateKey = {{.PrivateKey}}
Address ={{.PublicIp}}
DNS=10.43.0.10

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

func (wgc *wgManager) IsSetupDone() bool {
	_, err := wgc.remoteClient.Readfile(fmt.Sprintf("%v/config.json", wgc.configDataPath))
	if err != nil {
		return false
	}
	return true
}

func (wgc *wgManager) Init(ip string) (string, error) {
	fmt.Println("Initializing wireguard")

	out, err := wgc.remoteClient.Readfile(fmt.Sprintf("%v/config.json", wgc.configDataPath))

	var c Config

	if err != nil {
		fmt.Println("config.json not found generating new")

		key, err := wgtypes.GenerateKey()

		if err != nil {
			return "", fmt.Errorf("failed to generate public, private keys: %v", err)
		}

		nodeIp, err := wgc.GetNodeIp()
		c = Config{
			PublicKey:    key.PublicKey().String(),
			PrivateKey:   key.String(),
			PublicIp:     nodeIp,
			Peers:        map[string]Peer{},
			NetInterface: "eth0",
		}
	} else {
		fmt.Println("config found using old remote peers")
		err := json.Unmarshal([]byte(out), &c)

		if err != nil {
			return "", fmt.Errorf("failed to marshal config.json: %v", err)
		}

		c.Peers = map[string]Peer{}

	}

	var marshal []byte
	if marshal, err = json.Marshal(c); err != nil {
		fmt.Printf("failed to marshal config: %v\n", err)
		return "", fmt.Errorf("failed to masrshal config: %v ", err)
	}

	if err := wgc.remoteClient.WriteFile(fmt.Sprintf("%v/config.json", wgc.configDataPath), marshal); err != nil {
		return "", fmt.Errorf("failed to write config: %v ", err)
	}

	if err := c.writeConfig(*wgc); err != nil {
		panic(err)
		return "", fmt.Errorf("unable to write config file: %v", err)
	}

	return c.PublicKey, nil

}

func (wgc *wgManager) GetNodeIp() (string, error) {
	out, err := wgc.remoteClient.Readfile("wg-ip")
	if err != nil {
		return "10.13.13.1", nil
	}
	return string(out), nil
}

func (wgc *wgManager) AddRemotePeer(publicKey string, allowedIps string, endpoint *string) error {

	var c Config
	var configsRaw []byte
	var err error

	if configsRaw, err = wgc.remoteClient.Readfile(fmt.Sprintf("%v/config.json", wgc.configDataPath)); err != nil {
		return fmt.Errorf("unable read config.json: %v", err)
	}

	if err := json.Unmarshal(configsRaw, &c); err != nil {
		return fmt.Errorf("unable to parse config error: %v", err)
	}

	if c.RemotePeers == nil {
		c.RemotePeers = map[string]Peer{}
	}

	fmt.Println(c.RemotePeers, "remote peers")

	c.RemotePeers[publicKey] = Peer{
		PublicKey:  publicKey,
		Endpoint:   endpoint,
		AllowedIps: allowedIps,
	}
	err = c.writeConfig(*wgc)
	if err != nil {
		panic(err)
		return fmt.Errorf("unable to write config file: %v", err)
	}
	marshal, err := json.Marshal(c)
	err = wgc.remoteClient.WriteFile(fmt.Sprintf("%v/config.json", wgc.configDataPath), marshal)
	return err
}

func (wgc *wgManager) AddPeer(publicKey string, allowedIps string, endpoint *string) error {
	var c Config
	var configsRaw []byte
	var err error
	if configsRaw, err = wgc.remoteClient.Readfile(fmt.Sprintf("%v/config.json", wgc.configDataPath)); err != nil {
		return fmt.Errorf("unable read config.json: %v", err)
	}

	if err := json.Unmarshal(configsRaw, &c); err != nil {
		return fmt.Errorf("unable to parse config error: %v", err)
	}
	c.Peers[publicKey] = Peer{
		PublicKey:  publicKey,
		Endpoint:   endpoint,
		AllowedIps: allowedIps,
	}
	err = c.writeConfig(*wgc)
	if err != nil {
		panic(err)
		return fmt.Errorf("unable to write config file: %v", err)
	}
	marshal, err := json.Marshal(c)
	err = wgc.remoteClient.WriteFile(fmt.Sprintf("%v/config.json", wgc.configDataPath), marshal)
	return err
}

func (wgc *wgManager) getConfig() (Config, error) {
	var c Config
	var configsRaw []byte
	var err error
	if configsRaw, err = wgc.remoteClient.Readfile("config.json"); err != nil {
		return c, fmt.Errorf("unable to parse config error: %v", err)
	}
	if err := json.Unmarshal(configsRaw, &c); err != nil {
		return c, fmt.Errorf("unable to parse config error: %v", err)
	}
	return c, nil
}

func (wgc *wgManager) DeleteRemotePeer(publicKey string) error {
	var c Config
	var configsRaw []byte
	var err error
	if configsRaw, err = wgc.remoteClient.Readfile(fmt.Sprintf("%v/config.json", wgc.configDataPath)); err != nil {
		return fmt.Errorf("unable to parse config error: %v", err)
	}
	if err := json.Unmarshal(configsRaw, &c); err != nil {
		return fmt.Errorf("unable to parse config error: %v", err)
	}
	delete(c.RemotePeers, publicKey)
	return c.writeConfig(*wgc)
}

func (wgc *wgManager) DeletePeer(publicKey string) error {
	var c Config
	var configsRaw []byte
	var err error
	if configsRaw, err = wgc.remoteClient.Readfile(fmt.Sprintf("%v/config.json", wgc.configDataPath)); err != nil {
		return fmt.Errorf("unable to parse config error: %v", err)
	}
	if err := json.Unmarshal(configsRaw, &c); err != nil {
		return fmt.Errorf("unable to parse config error: %v", err)
	}
	delete(c.Peers, publicKey)
	return c.writeConfig(*wgc)
}
