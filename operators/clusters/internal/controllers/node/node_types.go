package node

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
