package target_node

import (
	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"
)

type AwsProviderConfig struct {
	AccessKey    string `yaml:"accessKey" json:"accessKey"`
	AccessSecret string `yaml:"accessSecret" json:"accessSecret"`
	AccountName  string `yaml:"accountName" json:"accountName"`
}

type CommonProviderData struct {
	TfTemplates string            `yaml:"tfTemplates"`
	Labels      map[string]string `yaml:"labels"`
	Taints      []string          `yaml:"taints"`
	SSHPath     string            `yaml:"sshPath"`
}

type AWSNodeConfig struct {
	clustersv1.AWSNodeConfig `json:",inline"`
	NodeName                 string `json:"nodeName"`
}
