package templates

import (
	"embed"

	"github.com/kloudlite/operator/pkg/templates"
)

//go:embed *
var templatesDir embed.FS

type templateFile string

const (
	// IngressTemplate templateFile = "./ingress-resource-v2.yml.tpl"
	IngressTemplate templateFile = "ingress-resource.yml.tpl"
)

func Read(t templateFile) ([]byte, error) {
	return templatesDir.ReadFile(string(t))
}

var ParseBytes = templates.ParseBytes
