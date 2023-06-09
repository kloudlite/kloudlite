package do

type DoProviderConfig struct {
	ApiToken  string `yaml:"apiToken"`
	AccountId string `yaml:"accountId"`
}

type DoNode struct {
	Region  string `yaml:"region"`
	Size    string `yaml:"size"`
	NodeId  string `yaml:"nodeId"`
	ImageId string `yaml:"imageId"`
}

type doClient struct {
	node DoNode

	apiToken string

	SSHPath     string
	accountId   string
	tfTemplates string
	labels      map[string]string
	taints      []string
}
