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
	VMJobTemplate templateFile = "./vm-job.yml.tpl"

// your entries
)

func Read(t templateFile) ([]byte, error) {
	return templatesDir.ReadFile(filepath.Join(string(t)))
}

var ParseBytes = templates.ParseBytes
