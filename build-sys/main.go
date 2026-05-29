package main

// Kloudlite provides Dagger functions for building kloudlite binaries, Docker images,
// CRDs, web apps, and orchestrating CI/CD pipelines.
type Kloudlite struct{}

// App is the component identifier for build+deploy pipelines.
// +enum
type App string

const (
	// +enum="api-server"
	AppAPIServer App = "api-server"
	// +enum="workmachine-manager"
	AppWorkmachineManager App = "workmachine-manager"
	// +enum="dashboard"
	AppDashboard App = "dashboard"
	// +enum="console"
	AppConsole App = "console"
	// +enum="website"
	AppWebsite App = "website"
	// +enum="oci-installer"
	AppOciInstaller App = "oci-installer"
	// +enum="code-analyzer"
	AppCodeAnalyzer App = "code-analyzer"
	// +enum="k3s-backup"
	AppK3sBackup App = "k3s-backup"
	// +enum="kloudlite"
	AppKloudlite App = "kloudlite"
	// +enum="kli"
	AppKli App = "kli"
	// +enum="kltun"
	AppKltun App = "kltun"
	// +enum="tunnel-server"
	AppTunnelServer App = "tunnel-server"
	// +enum="workspace-base"
	AppWorkspaceBase App = "workspace-base"
	// +enum="workspace-comprehensive"
	AppWorkspaceComprehensive App = "workspace-comprehensive"
)

// appImageName maps App enum to the image name suffix used in the registry.
func appImageName(app App) string {
	m := map[App]string{
		AppAPIServer:              "api-server",
		AppWorkmachineManager:     "workmachine-manager",
		AppDashboard:              "dashboard",
		AppConsole:                "console",
		AppWebsite:                "website",
		AppOciInstaller:           "oci-installer",
		AppCodeAnalyzer:           "code-analyzer",
		AppK3sBackup:              "k3s-backup",
		AppKloudlite:              "kloudlite",
		AppKli:                    "kli",
		AppKltun:                  "kltun",
		AppTunnelServer:           "tunnel-server",
		AppWorkspaceBase:          "workspace-base",
		AppWorkspaceComprehensive: "workspace-comprehensive",
	}
	if v, ok := m[app]; ok {
		return v
	}
	return string(app)
}

// appYamlManifest returns the manifest filename (without .yaml) for deployable apps.
func appYamlManifest(app App) string {
	switch app {
	case AppAPIServer:
		return "api-server"
	case AppDashboard:
		return "dashboard"
	}
	return ""
}
