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
	Project Project    `json:"project"`
	Entries []CCMEntry `json:"entries"`
}

type CCMEntry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Secret struct {
	Name    string     `json:"name"`
	Project Project    `json:"project"`
	Entries []CCMEntry `json:"entries"`
}

type JobVars struct {
	Name            string
	ServiceAccount  string
	Image           string
	ImagePullPolicy string
	Args            []string
	Env             map[string]string
}

type MsvcTemplate struct {
	Id         string `json:"id"`
	Name       string `json:"name"`
	Operations struct {
		Create string `json:"create"`
		Update string `json:"update"`
		Delete string `json:"delete"`
	} `json:"operations"`
	Resources []MsvcTemplateResource `json:"resources"`
}

type MsvcTemplateResource struct {
	Name         string `json:"name"`
	ResourceName string `json:"resourceName"`
	Operations   struct {
		Create string `json:"create"`
		Update string `json:"delete"`
		Delete string `json:"update"`
	} `json:"operations"`
}

type DomainSvc interface {
	ApplyJob(job *Job) error
}

type Job struct {
	Id      string `json:"id"`
	Project struct {
		Name string `json:"name"`
	} `json:"project"`

	Actions []JobAction
}

type JobAction struct {
	Id           string                 `json:"id"`
	ResourceType string                 `json:"resourceType"`
	ResourceId   string                 `json:"resourceId"`
	Action       string                 `json:"action"`
	Data         map[string]interface{} `json:"data"`
	KubeData     map[string]interface{} `json:"kubeData"`
	Flows        []string               `json:"flows"`
}
