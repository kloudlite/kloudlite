package server

import (
	"fmt"
	"os"

	"github.com/kloudlite/kl/domain/client"
	proxy "github.com/kloudlite/kl/domain/dev-proxy"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/fwd"
	"github.com/kloudlite/kl/pkg/ui/fzf"
)

var PaginationDefault = map[string]any{
	"orderBy":       "name",
	"sortDirection": "ASC",
	"first":         99999999,
}

// intercept {
//   enabled
//   toDevice
//   portMappings {
//     appPort
//     devicePort
//   }
// }
//

type AppSpec struct {
	Services []struct {
		Port int `json:"port"`
	} `json:"services"`
	Intercept struct {
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

func ListApps(options ...fn.Option) ([]App, error) {

	envName := fn.GetOption(options, "envName")

	env, err := EnsureEnv(&client.Env{
		Name: envName,
	}, options...)
	if err != nil {
		return nil, err
	}

	cookie, err := getCookie()
	if err != nil {
		return nil, err
	}

	respData, err := klFetch("cli_listApps", map[string]any{
		"pq":      PaginationDefault,
		"envName": env.Name,
	}, &cookie)

	if err != nil {
		return nil, err
	}

	if fromResp, err := GetFromRespForEdge[App](respData); err != nil {
		return nil, err
	} else {
		return fromResp, nil
	}
}

func SelectApp(options ...fn.Option) (*App, error) {

	appName := fn.GetOption(options, "appName")

	a, err := ListApps(options...)
	if err != nil {
		return nil, err
	}

	if len(a) == 0 {
		return nil, fmt.Errorf("no app found")
	}

	if appName != "" {
		for i, a2 := range a {
			if a2.Metadata.Name == appName {
				return &a[i], nil
			}
		}

		return nil, fmt.Errorf("app not found")
	}

	app, err := fzf.FindOne(a, func(item App) string {
		return fmt.Sprintf("%s (%s)%s", item.DisplayName, item.Metadata.Name, func() string {
			if item.IsMainApp {
				return ""
			}

			return " [external]"
		}())
	}, fzf.WithPrompt("Select App>"))
	if err != nil {
		return nil, err
	}

	return app, nil
}

func EnsureApp(options ...fn.Option) (*App, error) {

	s, err := SelectApp(options...)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func InterceptApp(status bool, ports []AppPort, options ...fn.Option) error {

	devName := fn.GetOption(options, "deviceName")
	envName := fn.GetOption(options, "envName")

	var err error

	if envName == "" {
		env, err := EnsureEnv(nil, options...)
		if err != nil {
			return err
		}

		envName = env.Name
	}

	if devName == "" {
		ctx, err := client.GetDeviceContext()
		if err != nil {
			return err
		}

		if ctx.DeviceName == "" {
			return fmt.Errorf("device name is required")
		}

		devName = ctx.DeviceName
	}

	app, err := EnsureApp(options...)
	if err != nil {
		return err
	}

	cookie, err := getCookie()
	if err != nil {
		return err
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

	if err := func() error {
		sshPort, ok := os.LookupEnv("SSH_PORT")
		if ok {
			var prs []fwd.StartCh

			for _, v := range ports {
				prs = append(prs, fwd.StartCh{
					SshPort:    sshPort,
					RemotePort: fmt.Sprint(v.DevicePort),
					LocalPort:  fmt.Sprint(v.DevicePort),
				})
			}

			p, err := proxy.NewProxy(false)
			if err != nil {
				return err
			}

			if status {
				if _, err := p.AddFwd(prs); err != nil {
					fn.PrintError(err)
					return err
				}
				return nil
			}

			if _, err := p.RemoveFwd(prs); err != nil {
				return err
			}
		}
		return nil
	}(); err != nil {
		fn.PrintError(err)
	}

	if len(ports) == 0 {
		return fmt.Errorf("no ports provided to intercept")
	}

	query := "cli_interceptApp"
	if !app.IsMainApp {
		query = "cli_intercepExternalApp"
	}

	respData, err := klFetch(query, map[string]any{
		"appName":      app.Metadata.Name,
		"envName":      envName,
		"deviceName":   devName,
		"intercept":    status,
		"portMappings": ports,
	}, &cookie)

	if err != nil {
		return err
	}

	if _, err := GetFromResp[bool](respData); err != nil {
		return err
	} else {
		return nil
	}
}
