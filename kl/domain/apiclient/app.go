package apiclient

import (
	"fmt"

	"github.com/kloudlite/kl/domain/fileclient"
	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
)

var PaginationDefault = map[string]any{
	"orderBy":       "name",
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

func (apic *apiClient) ListApps(accountName string, envName string) ([]App, error) {
	cookie, err := getCookie(fn.MakeOption("accountName", accountName))
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
// 		return nil, fmt.Errorf("no app found")
// 	}

// 	if appName != "" {
// 		for i, a2 := range a {
// 			if a2.Metadata.Name == appName {
// 				return &a[i], nil
// 			}
// 		}

// 		return nil, fmt.Errorf("app not found")
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

	devName := fn.GetOption(options, "deviceName")
	accountName := fn.GetOption(options, "accountName")

	fc, err := fileclient.New()
	if err != nil {
		return functions.NewE(err)
	}

	if accountName == "" {
		kt, err := fc.GetKlFile("")
		if err != nil {
			return functions.NewE(err)
		}

		if kt.AccountName == "" {
			return fmt.Errorf("account name is required")
		}

		accountName = kt.AccountName
		options = append(options, fn.MakeOption("accountName", accountName))
	}

	if devName == "" {
		avc, err := fc.GetVpnAccountConfig(accountName)
		if err != nil {
			return functions.NewE(err)
		}

		if avc.DeviceName == "" {
			return fmt.Errorf("device name is required")
		}

		devName = avc.DeviceName
	}

	cookie, err := getCookie([]fn.Option{
		fn.MakeOption("accountName", accountName),
	}...)
	if err != nil {
		return functions.NewE(err)
	}

	if len(ports) == 0 {
		if len(app.Spec.Intercept.PortMappings) != 0 {
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

	// if err := func() error {
	// 	sshPort, ok := os.LookupEnv("SSH_PORT")
	// 	if ok {
	// 		// var prs []sshclient.StartCh

	// 		// for _, v := range ports {
	// 		// 	prs = append(prs, sshclient.StartCh{
	// 		// 		SshPort:    sshPort,
	// 		// 		RemotePort: fmt.Sprint(v.DevicePort),
	// 		// 		LocalPort:  fmt.Sprint(v.DevicePort),
	// 		// 	})
	// 		// }

	// 		// TODO: add forwarding logic here
	// 		// p, err := proxy.NewProxy(false)
	// 		// if err != nil {
	// 		// 	return functions.NewE(err)
	// 		// }
	// 		//
	// 		// if status {
	// 		// 	if _, err := p.AddFwd(prs); err != nil {
	// 		// 		fn.PrintError(err)
	// 		// 		return functions.NewE(err)
	// 		// 	}
	// 		// 	return nil
	// 		// }
	// 		//
	// 		// if _, err := p.RemoveFwd(prs); err != nil {
	// 		// 	return functions.NewE(err)
	// 		// }
	// 	}
	// 	return nil
	// }(); err != nil {
	// 	fn.PrintError(err)
	// }

	if len(ports) == 0 {
		return fmt.Errorf("no ports provided to intercept")
	}

	query := "cli_interceptApp"
	if !app.IsMainApp {
		query = "cli_interceptExternalApp"
	}

	respData, err := klFetch(query, map[string]any{
		"appName":      app.Metadata.Name,
		"envName":      envName,
		"deviceName":   devName,
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
