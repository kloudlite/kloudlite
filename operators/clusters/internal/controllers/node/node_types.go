package node

type AwsProviderConfig struct {
	AccessKey    string `yaml:"accessKey" json:"accessKey"`
	AccessSecret string `yaml:"accessSecret" json:"accessSecret"`
	AccountName  string `yaml:"accountName" json:"accountName"`
}

type AWSNode struct {
	NodeId       string `yaml:"nodeId" json:"nodeId"`
	Region       string `yaml:"region" json:"region"`
	InstanceType string `yaml:"instanceType" json:"instanceType"`
	VPC          string `yaml:"vpc" json:"vpc"`
	ImageId      string `yaml:"imageId" json:"imageId"`
	IsGpu        bool   `yaml:"isGpu" json:"isGpu"`
	NodeType     string `yaml:"nodeType" json:"nodeType"`
}

type CommonProviderData struct {
	TfTemplates string            `yaml:"tfTemplates"`
	Labels      map[string]string `yaml:"labels"`
	Taints      []string          `yaml:"taints"`
	SSHPath     string            `yaml:"sshPath"`
}
