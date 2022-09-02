package text_templates

import (
	"github.com/Masterminds/sprig/v3"
	"text/template"
)

func WithFunctions(t *template.Template) *template.Template {
	return t.Funcs(sprig.TxtFuncMap())
}
