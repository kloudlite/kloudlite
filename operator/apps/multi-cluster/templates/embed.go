package templates

import (
	"bytes"
	"embed"
	"text/template"

	text_templates "github.com/kloudlite/operator/apps/multi-cluster/mpkg/text-template"
)

type TemplateName string

const (
	ServerConfg TemplateName = "wg-server.conf.tpl"
	ClientConfg TemplateName = "wg-client.conf.tpl"
)

//go:embed *
var templatesDir embed.FS

var tmplBytesMap = make(map[string]*template.Template)

func ParseTemplate(name TemplateName, data any) ([]byte, error) {

	if _, ok := tmplBytesMap[string(name)]; !ok {
		tb, err := templatesDir.ReadFile(string(name))
		if err != nil {
			return nil, err
		}
		tpl := text_templates.WithFunctions(template.New(string(name)))

		tpl, err = tpl.Parse(string(tb))
		if err != nil {
			return nil, err
		}

		tmplBytesMap[string(name)] = tpl
	}

	tmpl := tmplBytesMap[string(name)]

	out := new(bytes.Buffer)
	err := tmpl.Execute(out, data)
	if err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}
