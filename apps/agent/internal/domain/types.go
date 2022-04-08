package domain

type Project struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Logo        string `json:"logo"`
}

type Config struct {
	Name      string                 `json:"name"`
	Namespace string                 `json:"namespace"`
	Data      map[string]interface{} `json:"data"`
}

type Secret struct {
	Name      string                 `json:"name"`
	Namespace string                 `json:"namespace"`
	Data      map[string]interface{} `json:"data"`
}

type ManagedSvc struct {
	Name         string                 `json:"name"`
	Namespace    string                 `json:"namespace"`
	TemplateName string                 `json:"templateName"`
	Version      uint16                 `json:"version"`
	Values       map[string]interface{} `json:"values"`
	LastApplied  map[string]interface{} `json:"lastApplied"`
}

type ManagedRes struct {
	Name        string                 `json:"name"`
	Namespace   string                 `json:"namespace"`
	Type        string                 `json:"type"`
	ManagedSvc  string                 `json:"managedSvc"`
	Version     uint16                 `json:"version"`
	LastApplied map[string]interface{} `json:"lastApplied"`
	Values      map[string]interface{} `json:"values"`
}

type Routes struct {
	Path string `json:"path"`
	App  string `json:"app"`
	Port uint16 `json:"port"`
}

type Router struct {
	Name      string   `json:"name"`
	Namespace string   `json:"namespace"`
	Domains   []string `json:"domains"`
	Routes    []Routes `json:"routes"`
}

type buildArg struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type PipelineGithub struct {
	InstallationId string `json:"installationId"`
	TokenId        string `json:"tokenId"`
}

type PipelineGitlab struct {
	TokenId string `json:"tokenId"`
}

type Pipeline struct {
	Name        string         `json:"name"`
	Namespace   string         `json:"namespace"`
	GitProvider string         `json:"gitProvider"`
	GitRepoUrl  string         `json:"gitRepoUrl"`
	GitRef      string         `json:"gitRef"`
	BuildArgs   []buildArg     `json:"buildArgs"`
	Dockerfile  string         `json:"dockerfile"`
	ContextDir  string         `json:"contextDir"`
	Github      PipelineGithub `json:"github"`
	Gitlab      PipelineGitlab `json:"gitlab"`
}

type Message struct {
	ResourceType string      `json:"resourceType"`
	Namespace    string      `json:"namespace"`
	Spec         interface{} `json:"spec"`
}
