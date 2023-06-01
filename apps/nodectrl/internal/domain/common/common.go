package common

type CommonProviderData struct {
	TfTemplates string            `yaml:"tfTemplates"`
	Labels      map[string]string `yaml:"labels"`
	Taints      []string          `yaml:"taints"`
	SSHPath     string            `yaml:"sshPath"`
}
