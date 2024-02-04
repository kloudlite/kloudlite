package templates

import (
	"embed"

	"github.com/kloudlite/operator/pkg/templates"
)

//go:embed *
var templatesDir embed.FS

func ReadIngressTemplate() ([]byte, error) {
	return templatesDir.ReadFile("ingress-resource.yml.tpl")
}

var ParseBytes = templates.ParseBytes
