package templates

import (
	"bytes"
	"text/template"

	"github.com/kloudlite/kloudlite/operator/toolkit/errors"
)

type TextTemplate struct {
	*template.Template
}

func NewTextTemplate(name string) *TextTemplate {
	t := template.New(name).Option("missingkey=error")
	t.Funcs(txtFuncs(t))
	return &TextTemplate{t}
}

func (t *TextTemplate) ParseBytes(b []byte, values any) ([]byte, error) {
	if _, err := t.Parse(string(b)); err != nil {
		return nil, err
	}
	out := new(bytes.Buffer)
	if err := t.ExecuteTemplate(out, t.Name(), values); err != nil {
		return nil, errors.NewEf(err, "could not execute template")
	}
	return out.Bytes(), nil
}
