package server

import (
	"fmt"
	"github.com/kloudlite/kl/domain/client"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/fzf"
)

var PaginationDefault = map[string]any{
	"orderBy":       "name",
	"sortDirection": "ASC",
	"first":         99999999,
}

type App struct {
	DisplayName string   `json:"displayName"`
	Metadata    Metadata `json:"metadata"`
	Status      Status   `json:"status"`
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

func SelectApp(options ...fn.Option) (*string, error) {

	a, err := ListApps(options...)
	if err != nil {
		return nil, err
	}

	if len(a) == 0 {
		return nil, fmt.Errorf("no app found")
	}

	app, err := fzf.FindOne(a, func(item App) string {
		return fmt.Sprintf("%s (%s)", item.DisplayName, item.Metadata.Name)
	}, fzf.WithPrompt("Select App>"))
	if err != nil {
		return nil, err
	}

	return &app.Metadata.Name, nil
}

func EnsureApp(options ...fn.Option) (*string, error) {

	appName := fn.GetOption(options, "appName")
	envName := fn.GetOption(options, "envName")

	env, err := EnsureEnv(&client.Env{
		Name: envName,
	}, options...)

	if err != nil {
		return nil, err
	}

	envName = env.Name

	if appName == "" {
		s, err := SelectApp(options...)
		if err != nil {
			return nil, err
		}

		appName = *s
	}

	return &appName, nil
}

func InterceptApp(status bool, ports []AppPort, options ...fn.Option) error {

	appName := fn.GetOption(options, "appName")
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

	if appName == "" {
		s, err := EnsureApp(options...)
		if err != nil {
			return err
		}

		appName = *s
	}

	cookie, err := getCookie()
	if err != nil {
		return err
	}

	respData, err := klFetch("cli_interceptApp", map[string]any{
		"appname":      appName,
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
