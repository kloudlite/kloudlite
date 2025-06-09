package templates

import (
	"embed"
)

//go:embed *
var templatesDir embed.FS

func ReadBuildJobTemplate() ([]byte, error) {
	return templatesDir.ReadFile("build-job.yml.tpl")
}
