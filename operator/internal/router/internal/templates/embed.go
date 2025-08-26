package templates

import (
	"embed"

	"github.com/kloudlite/kloudlite/operator/toolkit/templates"
)

//go:embed *
var templatesDir embed.FS

type templateFile string

const (
	IngressTemplate templateFile = "ingress-resource.yml.tpl"
)

func Read(t templateFile) ([]byte, error) {
	return templatesDir.ReadFile(string(t))
}

var ParseBytes = templates.ParseBytes
