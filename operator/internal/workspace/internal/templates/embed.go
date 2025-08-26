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
	StatefulSetTemplate templateFile = "./statefulset.yml.tpl"
	ServiceTemplate     templateFile = "./service.yml.tpl"
	RouterTemplate      templateFile = "./router.yml.tpl"
)

func Read(tPaths ...templateFile) ([]byte, error) {
	var data []byte
	for _, t := range tPaths {
		d, err := templatesDir.ReadFile(filepath.Join(string(t)))
		if err != nil {
			return nil, err
		}
		data = append(data, d...)
	}
	return data, nil
}

var ParseBytes = templates.ParseBytes
