package target_nodepool

type NodeInfo struct {
	Name     string `yaml:"name" json:"name"`
	Status   string `yaml:"status" json:"status"`
	MetaData string `yaml:"metadata" json:"metadata"`
	Message  string `yaml:"message" json:"message"`
}
