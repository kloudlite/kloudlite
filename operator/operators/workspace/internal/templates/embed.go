package templates

import (
	"embed"
	"path/filepath"

	"github.com/kloudlite/operator/pkg/templates"
)

//go:embed deployments/*
var templatesDir embed.FS

type templateFile string

const (
	WorkspaceSTSTemplate     templateFile = "./deployments/sts.yml.tpl"
	WorkspaceIngressTemplate templateFile = "./deployments/ingress.yml.tpl"
	WorkspaceServiceTemplate templateFile = "./deployments/service.yml.tpl"
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
