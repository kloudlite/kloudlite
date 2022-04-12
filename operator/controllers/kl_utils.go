package controllers

const (
	RESOURCE_CONFIG           = "config"
	RESOURCE_SECRET           = "secret"
	RESOURCE_APP              = "app"
	RESOURCE_PROJECT          = "project"
	RESOURCE_ROUTER           = "router"
	RESOURCE_MANAGED_SERVICE  = "mService"
	RESOURCE_MANAGED_RESOURCE = "mResource"
)

type ResourceImage struct {
	Name string
}

func (r *ResourceImage) getImage() string {
	switch r.Name {
	case RESOURCE_PROJECT:
		return "harbor.dev.madhouselabs.io/kloudlite/jobs/project:latest"

	case RESOURCE_APP:
		return "harbor.dev.madhouselabs.io/kloudlite/jobs/app:latest"

	case RESOURCE_ROUTER:
		return "harbor.dev.madhouselabs.io/kloudlite/jobs/router:latest"

	case RESOURCE_MANAGED_SERVICE:
		return ""

	case RESOURCE_MANAGED_RESOURCE:
		return ""
	default:
		return ""
	}
}

type MSvcTemplate struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Operations  struct {
		Create string `json:"create"`
		Update string `json:"update"`
		Delete string `json:"delete"`
	} `json:"operations"`
	Resources []MresTemplate `json:"resources"`
}

type MresTemplate struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Operations struct {
		Create string `json:"create"`
		Update string `json:"delete"`
		Delete string `json:"update"`
	} `json:"operations"`
}
