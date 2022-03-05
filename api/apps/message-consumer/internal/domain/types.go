package domain

type Account struct {
	Id string `json:"id"`
}

type Project struct {
	Id          string  `json:"id"`
	Name        string  `json:"name"`
	Account     Account `json:"account"`
	DisplayName string  `json:"displayName"`
	Logo        string  `json:"logo"`
	Cluster     string  `json:"cluster"`
	Description string  `json:"description"`
}

type ProjectValues struct {
	Name        string `json:"name"`
	AccountId   string `json:"accountId"`
	DisplayName string `json:"displayName"`
	Logo        string `json:"logo"`
	Cluster     string `json:"cluster"`
	Description string `json:"description"`
}

type App struct {
	Id         string `json:"id"`
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	Replicas   int    `json:"replicas"`
	Containers struct {
		Name            string `json:"name"`
		Image           string `json:"image"`
		ImagePullPolicy string `json:"imagePullPolicy"`
		Env             []struct {
			Key     string `json:"name"`
			Value   string `json:"value"`
			RefKey  string `json:"refKey"`
			Type    string `json:"type"`
			RefName string `json:"refName"`
		} `json:"env"`
	} `json:"containers"`
	Services struct {
		Type       string `json:"type"`
		Port       int    `json:"port"`
		TargetPort int    `json:"targetPort"`
	} `json:"services"`
}

type Config struct {
	Name    string     `json:"name"`
	Entries []CCMEntry `json:"entries"`
}

type CCMEntry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Secret struct {
	Name    string     `json:"name"`
	Project Project    `json:"project"`
	Entries []CCMEntry `json:"data"`
}

type JobVars struct {
	Name            string
	ServiceAccount  string
	Image           string
	ImagePullPolicy string
	Args            []string
	Env             map[string]string
}

type DomainSvc interface {
	ApplyProject(project *Project) error
	DeleteProject(project *Project) error

	ApplyApp(app *App) error
	DeleteApp(app *App) error

	ApplyConfig(config *Config, project *Project) error
	DeleteConfig(config *Config) error

	ApplySecret(secret *Secret) error
	DeleteSecret(secret *Secret) error
}
