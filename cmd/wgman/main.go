package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os/exec"
	"text/template"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type Peer struct {
	PublicKey  string  `json:"public_key"`
	Endpoint   *string `json:"endpoint,omitempty"`
	AllowedIps string  `json:"allowed_ips"`
}

func (p *Peer) getConf() string {
	return fmt.Sprintf(`

`, p.PublicKey, p.AllowedIps)
}

type Config struct {
	PublicKey    string          `json:"public_key"`
	PrivateKey   string          `json:"private_key"`
	PublicIp     string          `json:"public_ip"`
	Peers        map[string]Peer `json:"peers"`
	NetInterface string          `json:"net_interface"`
}

func (c *Config) writeConfig() error {
	config, err := c.getWgConfig()
	if err != nil {
		return err
	}
	err = exec.Command("wg-quick", "down", "wg0").Run()
	exec.Command("rm", "/etc/wireguard/wg0.conf").Run()
	err = ioutil.WriteFile("/etc/wireguard/wg0.conf", []byte(config), 0644)
	err = exec.Command("wg-quick", "up", "wg0").Run()
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

func main() {
	var ip, peersBase64, command string
	flag.StringVar(&ip, "ip", "", "public ip")
	flag.StringVar(&command, "command", "", "command")
	flag.StringVar(&peersBase64, "peers", "", "peers json in Base64")

	flag.Parse()

	switch command {
	case "init":
		key, err := wgtypes.GenerateKey()
		if err != nil {
			panic(fmt.Errorf("failed to generate public, private keys: %v", err))
		}
		c := Config{
			PublicKey:    key.PublicKey().String(),
			PrivateKey:   key.String(),
			PublicIp:     ip,
			Peers:        map[string]Peer{},
			NetInterface: "eth0",
		}
		marshal, err := json.Marshal(c)
		if err != nil {
			panic(fmt.Errorf("failed to marshal config: %v", err))
		}
		err = exec.Command("rm", "./wg0.conf").Run()
		err = ioutil.WriteFile("config.json", marshal, 0644)
		if err != nil {
			panic(fmt.Errorf("unable to write config file: %v", err))
		}
		out, err := json.Marshal(map[string]string{
			"public_key": key.PublicKey().String(),
		})
		if err != nil {
			panic(fmt.Errorf("unable to generate output: %v", err))
		}
		c.writeConfig()
		fmt.Println(string(out))
		break
	case "peers":

		// all, _ := io.ReadAll(os.Stdin)
		// os.Stdin.Close()
		all, _ := base64.StdEncoding.DecodeString(peersBase64)
		peers := make([]Peer, 0)
		err := json.Unmarshal(all, &peers)
		if err != nil {
			fmt.Println(fmt.Errorf("unable to parse peers error: %v", err))
		}
		var c Config
		configsRaw, err := ioutil.ReadFile("config.json")
		err = json.Unmarshal(configsRaw, &c)
		if err != nil {
			fmt.Println(fmt.Errorf("unable to parse config error: %v", err))
		}

		for k := range c.Peers {
			delete(c.Peers, k)
		}

		for _, p := range peers {
			c.Peers[p.PublicKey] = p
		}

		marshal, err := json.Marshal(c)
		if err != nil {
			panic(fmt.Errorf("failed to marshal config: %v", err))
		}

		err = c.writeConfig()

		if err != nil {
			panic(fmt.Errorf("unable to update wireguard: %v", err))
		}

		err = ioutil.WriteFile("config.json", marshal, 0644)

		if err != nil {
			panic(fmt.Errorf("unable to write config file: %v", err))
		}

		break
	}
}
