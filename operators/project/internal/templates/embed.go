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
	HelmIngressNginx templateFile = "./helm-ingress-nginx.yml.tpl"
	NamespaceRBAC    templateFile = "./namespace-rbac.yml.tpl"
)

func Read(t templateFile) ([]byte, error) {
	return templatesDir.ReadFile(filepath.Join(string(t)))
}

var ParseBytes = templates.ParseBytes
