package templates

import (
	"embed"

	"github.com/kloudlite/operator/pkg/templates"
)

//go:embed *
var templatesDir embed.FS

func ReadNodepoolJobTemplate() ([]byte, error) {
	return templatesDir.ReadFile("nodepool-job.yml.tpl")
}

var ParseBytes = templates.ParseBytes
