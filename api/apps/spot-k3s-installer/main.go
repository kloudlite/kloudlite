package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

type NodeConfig struct {
	ServerIP string            `yaml:"serverIp"`
	Token    string            `yaml:"token"`
	NodeName string            `yaml:"nodeName"`
	Taints   []string          `yaml:"taints"`
	Labels   map[string]string `yaml:"labels"`
}

func ColorText(text interface{}, code int) string {
	return fmt.Sprintf("\033[38;05;%dm%v\033[0m", code, text)
}

func ExecCmd(cmdString string, logStr string) error {
	r := csv.NewReader(strings.NewReader(cmdString))
	r.Comma = ' '
	cmdArr, err := r.Read()
	if err != nil {
		return err
	}

	if logStr != "" {
		fmt.Printf("[#] %s\n", logStr)
	} else {
		fmt.Printf("[#] %s\n", strings.Join(cmdArr, " "))
	}

	cmd := exec.Command(cmdArr[0], cmdArr[1:]...)
	cmd.Stderr = os.Stderr
	// cmd.Stdout = os.Stdout

	if err := cmd.Run(); err != nil {
		fmt.Printf("err occurred: %v\n", err.Error())
		return err
	}
	return nil
}

func main() {

	for {
		if err := Run(); err != nil {
			fmt.Println(ColorText(err.Error(), 1))
			fmt.Println(ColorText("will retry after 10 second", 2))
			time.Sleep(time.Second * 10)
			continue
		}

		fmt.Println(ColorText("Successfully Installed", 2))
		break
	}
}

func ExecCmdWithOutput(cmdString string, logStr string) ([]byte, error) {
	r := csv.NewReader(strings.NewReader(cmdString))
	r.Comma = ' '
	cmdArr, err := r.Read()
	if err != nil {
		return nil, err
	}

	if logStr != "" {
		fmt.Printf("[#] %s\n", logStr)
	} else {
		fmt.Printf("[#] %s\n", strings.Join(cmdArr, " "))
	}

	cmd := exec.Command(cmdArr[0], cmdArr[1:]...)
	cmd.Stderr = os.Stderr
	// cmd.Stdout = os.Stdout

	return cmd.Output()
}

func Run() error {
	b, err := os.ReadFile("/k3s/data.yaml")
	if err != nil {
		return err
	}

	var nc NodeConfig

	if err := yaml.Unmarshal(b, &nc); err != nil {
		return err
	}

	var ip string

	if ipBytes, err := ExecCmdWithOutput("curl ifconfig.me", "fetching ip"); err != nil {
		return err
	} else {
		ip = strings.TrimSpace(string(ipBytes))
	}

	defaultLables := map[string]string{
		"kloudlite.io/public-ip": string(ip),
		"kloudlite.io/node-name": nc.NodeName,
	}

	labels := func() []string {
		l := []string{}
		for k, v := range defaultLables {
			l = append(l, fmt.Sprintf("--node-label %s=%s", k, v))
		}

		for k, v := range nc.Labels {
			l = append(l, fmt.Sprintf("--node-label %s=%s", k, v))
		}
		return l
	}()

	cmd := fmt.Sprintf(
		"k3s agent --server https://%s:6443 --token=%s --node-external-ip %s --node-name %s %s",
		nc.ServerIP,
		strings.TrimSpace(string(nc.Token)),
		ip,
		fmt.Sprintf("kl-worker-%s", ip),
		strings.Join(labels, " "),
	)

	if err := ExecCmd(cmd, "installing K3s"); err != nil {
		return err
	}

	return nil
}
