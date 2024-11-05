package templates

import (
	"bytes"
	"embed"
	"fmt"
	"path/filepath"
	"text/template"
)

//go:embed *
var templatesDir embed.FS

type templateFile string

const (
// TODO: (user) add your template files here
)

func Read(t templateFile) ([]byte, error) {
	return templatesDir.ReadFile(filepath.Join(string(t)))
}

func ParseBytes(b []byte, values any) ([]byte, error) {
	t := template.New("parse-bytes")
	// t := template.New("parse-bytes")
	// t.Funcs(txtFuncs(t))
	if _, err := t.Parse(string(b)); err != nil {
		return nil, err
	}
	out := new(bytes.Buffer)
	if err := t.ExecuteTemplate(out, "parse-bytes", values); err != nil {
		return nil, fmt.Errorf("could not execute template, as err %w occurred", err)
	}
	return out.Bytes(), nil
}
