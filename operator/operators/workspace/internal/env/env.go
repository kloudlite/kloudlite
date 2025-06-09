package env

import (
	"github.com/codingconcepts/env"
)

type Env struct {
	MaxConcurrentReconciles int `env:"MAX_CONCURRENT_RECONCILES"`

	WorkspaceImageInitContainer   string `env:"WORKSPACE_IMAGE_INIT_CONTAINER" default:"ghcr.io/kloudlite/iac/workspace:latest"`
	WorkspaceImageSSH             string `env:"WORKSPACE_IMAGE_SSH" default:"ghcr.io/kloudlite/iac/workspace:latest"`
	WorkspaceImageTTYD            string `env:"WORKSPACE_IMAGE_TTYD" default:"ghcr.io/kloudlite/iac/ttyd:latest"`
	WorkspaceImageJupyterNotebook string `env:"WORKSPACE_IMAGE_JUPYTER_NOTEBOOK" default:"ghcr.io/kloudlite/iac/jupyter:latest"`
	WorkspaceImageCodeServer      string `env:"WORKSPACE_IMAGE_CODE_SERVER" default:"ghcr.io/kloudlite/iac/code-server:latest"`
	WorkspaceImageVscodeServer    string `env:"WORKSPACE_IMAGE_VSCODE_SERVER" default:"ghcr.io/kloudlite/iac/vscode-server:latest"`
}

func GetEnvOrDie() *Env {
	var ev Env
	if err := env.Set(&ev); err != nil {
		panic(err)
	}

	return &ev
}
