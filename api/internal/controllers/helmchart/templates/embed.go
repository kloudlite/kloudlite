package templates

import (
	_ "embed"
	"github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/template"
)

type templateFile []byte

func (t templateFile) Render(values any) ([]byte, error) {
	return template.Render(t, values)
}

var (
	//go:embed helm-install-job.yml.tpl
	HelmInstallJobTemplate templateFile

	//go:embed helm-uninstall-job.yml.tpl
	HelmUninstallJobTemplate templateFile
)
