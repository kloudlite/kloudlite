package templates

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"text/template"

	"github.com/Masterminds/sprig"
	"operators.kloudlite.io/lib/errors"
)

func UseTemplate(filename templateFile) (KlTemplate, error) {
	tPath := path.Join(os.Getenv("PWD"), fmt.Sprintf("lib/templates/%s", filename))
	t, err := template.New(filename.String()).Funcs(sprig.TxtFuncMap()).ParseFiles(tPath)
	if err != nil {
		return nil, errors.NewEf(err, "could not parse template %s", filename)
	}
	return &kt{t}, nil
}

type kt struct {
	*template.Template
}

func (kt *kt) WithValues(v interface{}) ([]byte, error) {
	w := new(bytes.Buffer)
	if err := kt.Execute(w, v); err != nil {
		return nil, errors.NewEf(err, "could not execute template")
	}
	return w.Bytes(), nil
}

type KlTemplate interface {
	WithValues(v interface{}) ([]byte, error)
}

type templateFile string

func (tf templateFile) String() string {
	return string(tf)
}

const (
	MongoDBStandalone       templateFile = "mongodb-helm-standalone.tmpl.yml"
	App                     templateFile = "app.tmpl.yml"
	Service                 templateFile = "service.tmpl.yml"
	MongoDBResourceDatabase templateFile = "mongodb-resource-database.tmpl.yml"
)
