package v1

const (
	ProjectDomain string = "kloudlite.io"
)

const (
	PlatformServiceNameKey   string = ProjectDomain + "/platform-svc.name"
	PlatformServicePluginGVK string = ProjectDomain + "/platform-svc.gvk"
)

const (
	EnvironmentNameKey string = ProjectDomain + "/environment.name"
)

const (
	AppNameKey string = ProjectDomain + "/app.name"
)

const (
	GatewayEnabledLabelKey   string = ProjectDomain + "/gateway.enabled"
	GatewayEnabledLabelValue string = "true"
)

const (
	WorkspaceNameKey   string = ProjectDomain + "/workspace.name"
	WorkMachineNameKey string = ProjectDomain + "/workmachine.name"
)
