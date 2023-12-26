package main

import "fmt"

type K3sCommonFlags struct {
	Token    string   `json:"token"`
	NodeName string   `json:"nodeName"`
	Taints   []string `json:"taints"`

	// INFO: k3s only supports node-labels, not node-annotations
	Labels map[string]string `json:"labels"`
}

func (kcf K3sCommonFlags) ParseIntoFlags() []string {
	flags := make([]string, 0, 2*(len(kcf.Taints)+len(kcf.Labels)))

	flags = append(flags, "--node-name", kcf.NodeName)
	flags = append(flags, "--token", kcf.Token)

	for i := range kcf.Taints {
		flags = append(flags, "--node-taint", kcf.Taints[i])
	}

	for k, v := range kcf.Labels {
		flags = append(flags, "--node-label", fmt.Sprintf("%s=%s", k, v))
	}

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
