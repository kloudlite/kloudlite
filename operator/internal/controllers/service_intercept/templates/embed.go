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
	SvcIntercept    templateFile = "./svc-intercept.yml.tpl"
	WebhookTemplate templateFile = "./webhook.yml.tpl"
)

func Read(t templateFile) ([]byte, error) {
	return templatesDir.ReadFile(filepath.Join(string(t)))
}

var ParseBytes = templates.ParseBytes

