package templates

import (
	"embed"
	"path/filepath"

	"github.com/kloudlite/kloudlite/operator/toolkit/templates"
)

//go:embed *
var templatesDir embed.FS

type templateFile string

const (
	HPASpec             templateFile = "./hpa-spec.yml.tpl"
	AppInterceptPodSpec templateFile = "./app-intercept-pod-spec.yml.tpl"
	Deployment          templateFile = "./deployment.yml.tpl"
)

func Read(t templateFile) ([]byte, error) {
	return templatesDir.ReadFile(filepath.Join(string(t)))
}

var ParseBytes = templates.ParseBytes
