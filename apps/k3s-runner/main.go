package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/kloudlite/operator/pkg/constants"
	"sigs.k8s.io/yaml"
)

type K3sCommonFlags struct {
	Token       string            `json:"token"`
	NodeName    string            `json:"nodeName"`
	Taints      []string          `json:"taints"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}

func (kcf K3sCommonFlags) ParseIntoFlags() []string {
	flags := make([]string, 0, 2*(len(kcf.Taints)+len(kcf.Labels)+len(kcf.Annotations)+2))

	flags = append(flags, "--node-name", kcf.NodeName)
	flags = append(flags, "--token", kcf.Token)

	for i := range kcf.Taints {
		flags = append(flags, "--node-taint", kcf.Taints[i])
	}

	for k, v := range kcf.Labels {
		flags = append(flags, "--node-label", fmt.Sprintf("%s=%s", k, v))
	}

	// INFO: k3s does not support node-annotation yet
	// for k, v := range kcf.Annotations {
	// 	flags = append(flags, "--node-annotation", fmt.Sprintf("%s=%s", k, v))
	// }
	return flags
}

type AgentConfig struct {
	PublicIP       string `json:"publicIP"`
	ServerIP       string `json:"serverIP"`
	K3sCommonFlags `json:",inline"`
	ExtraAgentArgs []string `json:"extraAgentArgs"`
}

type PrimaryMasterConfig struct {
	PublicIP        string   `json:"publicIP"`
	SANs            []string `json:"SANs"`
	K3sCommonFlags  `json:",inline"`
	ExtraServerArgs []string `json:"extraServerArgs"`
}

type SecondaryMasterConfig struct {
	PublicIP        string   `json:"publicIP"`
	ServerIP        string   `json:"serverIP"`
	SANs            []string `json:"SANs"`
	K3sCommonFlags  `json:",inline"`
	ExtraServerArgs []string `json:"extraServerArgs"`
}

type RunAsMode string

const (
	RunAsAgent           RunAsMode = "agent"
	RunAsPrimaryMaster   RunAsMode = "primaryMaster"
	RunAsSecondaryMaster RunAsMode = "secondaryMaster"
)

type K3sRunnerConfig struct {
	RunAs           RunAsMode              `json:"runAs"`
	Agent           *AgentConfig           `json:"agent"`
	PrimaryMaster   *PrimaryMasterConfig   `json:"primaryMaster"`
	SecondaryMaster *SecondaryMasterConfig `json:"secondaryMaster"`
}

func ColorText(text interface{}, code int) string {
	return fmt.Sprintf("\033[38;05;%dm%v\033[0m", code, text)
}

func execK3s(ctx context.Context, args ...string) error {
	stdout, err := os.Create("runner.stdout.log")
	if err != nil {
		return err
	}

	stderr, err := os.Create("runner.stderr.log")
	if err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, "k3s", args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	fmt.Fprintf(stdout, "executing this shell command: %s\n", cmd.String())
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(stdout, "[ERROR]: %s", err.Error())
		return err
	}
	return nil
}

func getPublicIPv4() (string, error) {
	req, err := http.NewRequest(http.MethodGet, "https://ifconfig.me", nil)
	if err != nil {
		return "", err
	}

	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	b, err := io.ReadAll(r.Body)
	if err != nil {
		return "", err
	}
	defer r.Body.Close()

	return string(b), nil
}

func main() {
	runnerCfgFile := "/runner-config.yml"

	ctx, cf := signal.NotifyContext(context.TODO(), syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	defer cf()

	for {
		if err := ctx.Err(); err != nil {
			fmt.Println("context cancelled")
			os.Exit(1)
		}
		f, err := os.Open(runnerCfgFile)
		if err != nil {
			fmt.Println(ColorText(err.Error(), 1))
			continue
		}

		fmt.Println("found runner config file")

		out, err := io.ReadAll(f)
		if err != nil {
			fmt.Println(ColorText(err.Error(), 1))
			continue
		}

		var runnerCfg K3sRunnerConfig
		if err := yaml.Unmarshal(out, &runnerCfg); err != nil {
			fmt.Println(ColorText(err.Error(), 1))
			continue
		}

		switch runnerCfg.RunAs {
		case RunAsAgent:
			{
				if err := StartK3sAgent(ctx, runnerCfg.Agent); err != nil {
					if !errors.Is(err, context.Canceled) {
						fmt.Println(ColorText(err.Error(), 1))
						fmt.Println(ColorText("will retry after 10 second", 2))
						time.Sleep(time.Second * 10)
					}
				}
			}

		case RunAsPrimaryMaster:
			{
				if err := StartPrimaryK3sMaster(ctx, runnerCfg.PrimaryMaster); err != nil {
					if !errors.Is(err, context.Canceled) {
						fmt.Println(ColorText(err.Error(), 1))
						fmt.Println(ColorText("will retry after 10 second", 2))
						time.Sleep(time.Second * 10)
					}
				}
			}

		case RunAsSecondaryMaster:
			{
				if err := StartSecondaryK3sMaster(ctx, runnerCfg.SecondaryMaster); err != nil {
					if !errors.Is(err, context.Canceled) {
						fmt.Println(ColorText(err.Error(), 1))
						fmt.Println(ColorText("will retry after 10 second", 2))
						time.Sleep(time.Second * 10)
					}
				}
			}
		default:
			{
				fmt.Println(ColorText("invalid runAs mode", 1))
				continue
			}
		}

		fmt.Println(ColorText("Successfully Installed", 2))
		break
	}
}

func StartPrimaryK3sMaster(ctx context.Context, pmc *PrimaryMasterConfig) error {
	fmt.Printf("starting as primary master, with configuration: %#v\n", *pmc)

	argsAndFlags := []string{
		"server",
		"--cluster-init",
		"--flannel-backend", "wireguard-native",
		"--write-kubeconfig-mode", "644",
		"--flannel-backend", "wireguard-native",
		"--node-label", fmt.Sprintf("%s=%s", constants.PublicIpKey, pmc.PublicIP),
		"--node-label", fmt.Sprintf("%s=%s", constants.NodeNameKey, pmc.NodeName),
		"--tls-san", pmc.PublicIP,
	}

	for i := range pmc.SANs {
		argsAndFlags = append(argsAndFlags, "--tls-san", pmc.SANs[i])
	}

	argsAndFlags = append(argsAndFlags, pmc.K3sCommonFlags.ParseIntoFlags()...)

	argsAndFlags = append(argsAndFlags, pmc.ExtraServerArgs...)

	return execK3s(ctx, argsAndFlags...)
}

func StartSecondaryK3sMaster(ctx context.Context, smc *SecondaryMasterConfig) error {
	argsAndFlags := []string{
		"server",
		"--server", fmt.Sprintf("https://%s:6443", smc.ServerIP),
		"--flannel-backend", "wireguard-native",
		"--write-kubeconfig-mode", "644",
		"--node-label", fmt.Sprintf("%s=%s", constants.PublicIpKey, smc.PublicIP),
		"--node-label", fmt.Sprintf("%s=%s", constants.NodeNameKey, smc.NodeName),
		"--tls-san", smc.PublicIP,
	}

	for i := range smc.SANs {
		argsAndFlags = append(argsAndFlags, "--tls-san", smc.SANs[i])
	}

	argsAndFlags = append(argsAndFlags, smc.K3sCommonFlags.ParseIntoFlags()...)

	argsAndFlags = append(argsAndFlags, smc.ExtraServerArgs...)

	return execK3s(ctx, argsAndFlags...)
}

func StartK3sAgent(ctx context.Context, agentCfg *AgentConfig) error {
	ip, err := func() (string, error) {
		if agentCfg.PublicIP != "" {
			return agentCfg.PublicIP, nil
		}

		return getPublicIPv4()
	}()

	if err != nil {
		return err
	}

	argsAndFlags := []string{
		"agent",
		"--server", fmt.Sprintf("https://%s:6443", agentCfg.ServerIP),
		"--node-label", fmt.Sprintf("%s=%s", constants.PublicIpKey, ip),
		"--node-label", fmt.Sprintf("%s=%s", constants.NodeNameKey, agentCfg.NodeName),
	}

	argsAndFlags = append(argsAndFlags, agentCfg.K3sCommonFlags.ParseIntoFlags()...)

	argsAndFlags = append(argsAndFlags, agentCfg.ExtraAgentArgs...)

	return execK3s(ctx, argsAndFlags...)
}
