package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/kloudlite/operator/pkg/constants"

	"sigs.k8s.io/yaml"
)

var logger = log.New(os.Stdout, "k3s-runner", log.LstdFlags)

type K3sRunnerConfig struct {
	RunAs           RunAsMode              `json:"runAs"`
	Agent           *AgentConfig           `json:"agent"`
	PrimaryMaster   *PrimaryMasterConfig   `json:"primaryMaster"`
	SecondaryMaster *SecondaryMasterConfig `json:"secondaryMaster"`
}

func execK3s(ctx context.Context, args ...string) error {
	cmd := exec.CommandContext(ctx, "k3s", args...)
	logger.Printf("executing shell cmd: %s\n", cmd.String())

	if err := cmd.Run(); err != nil {
		logger.Printf("encountered error: %v\n", err)
		return err
	}

	return cmd.Wait()
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
	var runnerCfgFile string
	flag.StringVar(&runnerCfgFile, "config", "", "--config runner-config-file")
	flag.Parse()

	logger.Printf("specified configuration file: %s\n", runnerCfgFile)

	ctx, cf := signal.NotifyContext(context.TODO(), syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	defer cf()

	for {
		if err := ctx.Err(); err != nil {
			fmt.Println("context cancelled")
			os.Exit(1)
		}
		f, err := os.Open(runnerCfgFile)
		if err != nil {
			logger.Printf("failed to open configuration file, encountered err: %v", err)
			continue
		}

		fmt.Println("found runner config file")

		out, err := io.ReadAll(f)
		if err != nil {
			logger.Printf("failed to read configuration file, encountered err: %v", err)
			continue
		}

		var runnerCfg K3sRunnerConfig
		if err := yaml.Unmarshal(out, &runnerCfg); err != nil {
			logger.Printf("encountered err, while unmarshalling: %v", err)
			continue
		}

		switch runnerCfg.RunAs {
		case RunAsAgent:
			{
				if err := StartK3sAgent(ctx, runnerCfg.Agent); err != nil {
					if !errors.Is(err, context.Canceled) {
						logger.Printf("failed to start k3s in agent mode, encountered err: %v", err)
						logger.Printf("will retry in 10 seconds")
						<-time.After(10 * time.Second)
						continue
					}
				}
			}

		case RunAsPrimaryMaster:
			{
				if err := StartPrimaryK3sMaster(ctx, runnerCfg.PrimaryMaster); err != nil {
					if !errors.Is(err, context.Canceled) {
						logger.Printf("failed to start k3s in server mode (primary), encountered err: %v", err)
						logger.Printf("will retry in 10 seconds")
						<-time.After(10 * time.Second)
						continue
					}
				}
			}

		case RunAsSecondaryMaster:
			{
				if err := StartSecondaryK3sMaster(ctx, runnerCfg.SecondaryMaster); err != nil {
					if !errors.Is(err, context.Canceled) {
						logger.Printf("failed to start k3s in server mode (secondary), encountered err: %v", err)
						logger.Printf("will retry in 10 seconds")
						<-time.After(10 * time.Second)
						continue
					}
				}
			}
		default:
			{
				logger.Printf("invalid runAs mode: %v", runnerCfg.RunAs)
				continue
			}
		}

		logger.Printf("successfully started runner in %s mode", runnerCfg.RunAs)
		break
	}
}

func StartPrimaryK3sMaster(ctx context.Context, pmc *PrimaryMasterConfig) error {
	logger.Printf("starting runner as primary master, with configuration: %#v\n", *pmc)

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

	argsAndFlags = append(argsAndFlags, pmc.ParseIntoFlags()...)

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

	argsAndFlags = append(argsAndFlags, agentCfg.ParseIntoFlags()...)

	argsAndFlags = append(argsAndFlags, agentCfg.ExtraAgentArgs...)

	return execK3s(ctx, argsAndFlags...)
}
