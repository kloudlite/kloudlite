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

func UseTemplate(tfile templateFile, tList ...templateFile) (KlTemplate, error) {
	t := template.New(tfile.String())
	tList = append(tList, tfile)
	for _, f := range tList {
		templatesDir := path.Join(os.Getenv("PWD"), "lib/templates")
		if v, hasTemplatesDir := os.LookupEnv("TEMPLATES_DIR"); hasTemplatesDir {
			templatesDir = v
		}

		tPath := fmt.Sprintf("%s/%s", templatesDir, f.String())
		_, err := t.New(f.String()).Funcs(sprig.TxtFuncMap()).ParseFiles(tPath)
		if err != nil {
			return nil, errors.NewEf(err, "could not parse template %s", f.String())
		}
	}
	return &kt{list: tList, Template: t}, nil
}

type kt struct {
	list []templateFile
	*template.Template
}

func (kt *kt) WithValues(v interface{}) ([]byte, error) {
	w := new(bytes.Buffer)
	for _, t := range kt.list {
		if err := kt.ExecuteTemplate(w, t.String(), v); err != nil {
			return nil, errors.NewEf(err, "could not execute template")
		}
	}
	return w.Bytes(), nil
}

func Parse(f templateFile, values interface{}) ([]byte, error) {
	t, err := UseTemplate(f)
	if err != nil {
		return nil, err
	}
	return t.WithValues(values)
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
	MongoDBCluster          templateFile = "mongodb-helm-cluster.tmpl.yml"
	MongoDBWatcher          templateFile = "mongo-msvc-watcher.tmpl.yml"
	Deployment              templateFile = "deployment.tmpl.yml"
	Service                 templateFile = "service.tmpl.yml"
	Secret                  templateFile = "secret.tmpl.yml"
	MongoDBResourceDatabase templateFile = "mongodb-resource-database.tmpl.yml"
	AccountWireguard        templateFile = "account-deploy.tmpl.yml"
)
