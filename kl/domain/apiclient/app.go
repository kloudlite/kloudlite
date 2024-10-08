package apiclient

import (
	"fmt"
	"github.com/kloudlite/kl/domain/fileclient"
	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/spinner"
)

var PaginationDefault = map[string]any{
	"orderBy":       "updateTime",
	"sortDirection": "ASC",
	"first":         99999999,
}

type AppSpec struct {
	Services []struct {
		Port int `json:"port"`
	} `json:"services"`
	Intercept *struct {
		Enabled      bool      `json:"enabled"`
		PortMappings []AppPort `json:"portMappings"`
	} `json:"intercept"`
}

type App struct {
	DisplayName string   `json:"displayName"`
	Metadata    Metadata `json:"metadata"`
	Spec        AppSpec  `json:"spec"`
	Status      Status   `json:"status"`
	IsMainApp   bool     `json:"mapp"`
}

type AppPort struct {
	AppPort    int `json:"appPort"`
	DevicePort int `json:"devicePort,omitempty"`
}

func (apic *apiClient) ListApps(teamName string, envName string) ([]App, error) {
	cookie, err := getCookie(fn.MakeOption("teamName", teamName))
	if err != nil {
		return nil, functions.NewE(err)
	}
	respData, err := klFetch("cli_listApps", map[string]any{
		"pq":      PaginationDefault,
		"envName": envName,
	}, &cookie)
	if err != nil {
		return nil, functions.NewE(err)
	}
	if fromResp, err := GetFromRespForEdge[App](respData); err != nil {
		return nil, functions.NewE(err)
	} else {
		return fromResp, nil
	}
}

// func (apic *apiClient) SelectApp(options ...fn.Option) (*App, error) {
// 	appName := fn.GetOption(options, "appName")

// 	a, err := apic.ListApps(options...)
// 	if err != nil {
// 		return nil, functions.NewE(err)
// 	}

// 	if len(a) == 0 {
// 		return nil, fn.Errorf("no app found")
// 	}

// 	if appName != "" {
// 		for i, a2 := range a {
// 			if a2.Metadata.Name == appName {
// 				return &a[i], nil
// 			}
// 		}

// 		return nil, fn.Errorf("app not found")
// 	}

// 	app, err := fzf.FindOne(a, func(item App) string {
// 		return fmt.Sprintf("%s (%s)%s", item.DisplayName, item.Metadata.Name, func() string {
// 			if item.IsMainApp {
// 				return ""
// 			}

// 			return " [external]"
// 		}())
// 	}, fzf.WithPrompt("Select App>"))
// 	if err != nil {
// 		return nil, functions.NewE(err)
// 	}

// 	return app, nil
// }

// func EnsureApp(envName string, options ...fn.Option) (*App, error) {

// 	s, err := SelectApp(envName, options...)
// 	if err != nil {
// 		return nil, functions.NewE(err)
// 	}

// 	return s, nil
// }

func (apic *apiClient) InterceptApp(app *App, status bool, ports []AppPort, envName string, options ...fn.Option) error {
	teamName := fn.GetOption(options, "teamName")
	devName := fn.GetOption(options, "deviceName")

	fc, err := fileclient.New()
	if err != nil {
		return functions.NewE(err)
	}

	if teamName == "" {
		kt, err := fc.GetKlFile("")
		if err != nil {
			return functions.NewE(err)
		}

		if kt.TeamName == "" {
			return fn.Errorf("team name is required")
		}

		teamName = kt.TeamName
		options = append(options, fn.MakeOption("teamName", teamName))
	}

	if devName == "" {
		avc, err := fc.GetVpnTeamConfig(teamName)
		if err != nil {
			return functions.NewE(err)
		}

		if avc.DeviceName == "" {
			return fmt.Errorf("device name is required")
		}

		devName = avc.DeviceName
	}

	cookie, err := getCookie([]fn.Option{
		fn.MakeOption("teamName", teamName),
	}...)
	if err != nil {
		return functions.NewE(err)
	}

	if len(ports) == 0 {
		if app.Spec.Intercept != nil && len(app.Spec.Intercept.PortMappings) != 0 {
			ports = append(ports, app.Spec.Intercept.PortMappings...)
		} else if len(app.Spec.Services) != 0 {
			for _, v := range app.Spec.Services {
				ports = append(ports, AppPort{
					AppPort:    v.Port,
					DevicePort: v.Port,
				})
			}
		}
	}

	if len(ports) == 0 {
		return fn.Errorf("no ports provided to intercept")
	}

	//user, err := apic.GetCurrentUser()
	//if err != nil {
	//	return err
	//}
	//
	//hostName := os.Getenv("KL_HOST_USER")

	query := "cli_interceptApp"
	if !app.IsMainApp {
		query = "cli_interceptExternalApp"
	}

	//k3sTracker, err := apic.fc.GetK3sTracker()
	//if err != nil {
	//	return fn.Error("k3s server is not ready, please wait")
	//}

	//lastCheckedAt, err := time.Parse(time.RFC3339, k3sTracker.LastCheckedAt)
	//if err != nil {
	//	return fn.Error("k3s server is not ready, please wait")
	//}
	//
	//if time.Since(lastCheckedAt) > 3*time.Second {
	//	return fn.Error("k3s server is not ready, please wait")
	//}

	respData, err := klFetch(query, map[string]any{
		"appName":    app.Metadata.Name,
		"envName":    envName,
		"deviceName": devName,
		//"ipAddr":       k3sTracker.DeviceRouter.IP,
		//"clusterName":  fmt.Sprintf("%s-%s", user.Name, hostName),
		"intercept":    status,
		"portMappings": ports,
	}, &cookie)
	if err != nil {
		return functions.NewE(err)
	}

	if _, err := GetFromResp[bool](respData); err != nil {
		return functions.NewE(err)
	} else {
		return nil
	}
}

func (apic *apiClient) RemoveAllIntercepts(options ...fn.Option) error {
	defer spinner.Client.UpdateMessage("Cleaning up intercepts...")()
	// devName := fn.GetOption(options, "deviceName")
	teamName := fn.GetOption(options, "teamName")
	currentEnv, err := apic.EnsureEnv()
	if err != nil {
		return functions.NewE(err)
	}

	fc, err := fileclient.New()
	if err != nil {
		return functions.NewE(err)
	}

	if teamName == "" {
		kt, err := fc.GetKlFile("")
		if err != nil {
			return functions.NewE(err)
		}

		if kt.TeamName == "" {
			return fn.Errorf("team name is required")
		}

		teamName = kt.TeamName
		options = append(options, fn.MakeOption("teamName", teamName))
	}

	config, err := apic.fc.GetClusterConfig(teamName)
	if err != nil {
		return functions.NewE(err)
	}

	//if devName == "" {
	//	avc, err := fc.GetVpnTeamConfig(teamName)
	//	if err != nil && os.IsNotExist(err) {
	//		return nil
	//	} else if err != nil {
	//		return functions.NewE(err)
	//	}
	//
	//	if avc.DeviceName == "" {
	//		return fn.Errorf("device name is required")
	//	}
	//
	//	devName = avc.DeviceName
	//}

	cookie, err := getCookie([]fn.Option{
		fn.MakeOption("teamName", teamName),
	}...)
	if err != nil {
		return functions.NewE(err)
	}
	query := "cli_removeDeviceIntercepts"

	respData, err := klFetch(query, map[string]any{
		"envName": currentEnv.Name,
		//"deviceName": devName,
		"deviceName": config.ClusterName,
	}, &cookie)
	if err != nil {
		return functions.NewE(err)
	}

	if _, err := GetFromResp[bool](respData); err != nil {
		return functions.NewE(err)
	} else {
		return nil
	}
}
