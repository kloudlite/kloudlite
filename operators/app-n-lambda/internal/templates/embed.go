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
	// TODO: (user) add your template files here
	// ClusterJobTemplate        templateFile = "./cluster-job.yml.tpl"
	AppDeployment templateFile = "./app-deployment-svc-hpa.yml.tpl"
	HPATemplate   templateFile = "./hpa-template.yml.tpl"
	AppIntercept templateFile = "./app-intercept.yml.tpl"
)

func Read(t templateFile) ([]byte, error) {
	return templatesDir.ReadFile(filepath.Join(string(t)))
}

var ParseBytes = templates.ParseBytes
