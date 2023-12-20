package text_templates

import (
	"bytes"
	"github.com/kloudlite/api/pkg/errors"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

func WithFunctions(t *template.Template) *template.Template {
	var funcs template.FuncMap = map[string]any{}
	funcs["include"] = func(templateName string, templateData any) (string, error) {
		buf := bytes.NewBuffer(nil)
		if err := t.ExecuteTemplate(buf, templateName, templateData); err != nil {
			return "", errors.NewE(err)
		}
		return buf.String(), nil
	}

	return t.Funcs(sprig.TxtFuncMap()).Funcs(funcs)
}
