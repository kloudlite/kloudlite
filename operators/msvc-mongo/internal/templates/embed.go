package templates

import (
	"embed"

	"github.com/kloudlite/operator/pkg/templates"
)

//go:embed *
var templatesDir embed.FS

var ParseBytes = templates.ParseBytes

func ReadHelmMongoDBClusterTemplate() ([]byte, error) {
	return templatesDir.ReadFile("helm-mongodb-cluster.yml.tpl")
}
