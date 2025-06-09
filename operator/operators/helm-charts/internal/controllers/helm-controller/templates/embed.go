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
	JobRBACTemplate                 templateFile = "./job-rbac.yml.tpl"
	HelmInstallOrUpgradeJobTemplate templateFile = "./install-or-upgrade-job.yml.tpl"
	HelmUninstallJobTemplate        templateFile = "./uninstall-job.yml.tpl"
)

func Read(t templateFile) ([]byte, error) {
	return templatesDir.ReadFile(filepath.Join(string(t)))
}

var ParseBytes = templates.ParseBytes
