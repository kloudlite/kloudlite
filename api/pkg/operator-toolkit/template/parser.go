package template

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

func newTemplate(name string) (*template.Template, error) {
	t := template.New(name).Option("missingkey=error")
	t.Funcs(sprig.TxtFuncMap())
	return t, nil
}

func Render(b []byte, values any) ([]byte, error) {
	t, err := newTemplate("render-bytes")
	if err != nil {
		return nil, err
	}
	if _, err := t.Parse(string(b)); err != nil {
		return nil, err
	}
	out := new(bytes.Buffer)
	if err := t.ExecuteTemplate(out, "render-bytes", values); err != nil {
		return nil, fmt.Errorf("failed to render template: %w", err)
	}
	return out.Bytes(), nil
}
