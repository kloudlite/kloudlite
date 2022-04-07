package domain

type Project struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
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
	Name      string                 `json:"name"`
	Namespace string                 `json:"namespace"`
	Data      map[string]interface{} `json:"data"`
}

type ManagedRes struct {
	Name       string                 `json:"name"`
	Namespace  string                 `json:"namespace"`
	ManagedSvc ManagedSvc             `json:"managed_svc"`
	Data       map[string]interface{} `json:"data"`
}

type Message struct {
	ResourceType string      `json:"resource_type"`
	Namespace    string      `json:"namespace"`
	Spec         interface{} `json:"spec"`
}
