package templates

import (
	"embed"
	"path/filepath"

	"github.com/kloudlite/operator/pkg/templates"
)

//go:embed *
var templatesDir embed.FS

type templateFile string

const (
	WorkspaceTemplate            templateFile = "./workspace.yml.tpl"
	WorkMachineLifecycleTemplate templateFile = "./workmachine-lifecycle.yml.tpl"
	JumpServerDeploymentSpec     templateFile = "./ssh-jumpserver-deployment-spec.yml.tpl"
	AuthorizedKeysTemplate       templateFile = "./ssh-jumpserver-authorized-keys.tpl"
)

func Read(t templateFile) ([]byte, error) {
	return templatesDir.ReadFile(filepath.Join(string(t)))
}

var ParseBytes = templates.ParseBytes
