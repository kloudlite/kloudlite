package templates

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"kloudlite.io/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/Masterminds/sprig/v3"
)

func useTemplate(file templateFile) (*kt, error) {
	tFiles := []string{file.Path()}
	tFiles = append(tFiles, helperFiles...)

	// SOURCE: https://github.com/helm/helm/blob/8648ccf5d35d682dcd5f7a9c2082f0aaf071e817/pkg/engine/engine.go#L147-L154
	t := template.New(filepath.Base(file.Path()))
	var funcMap template.FuncMap = map[string]any{}
	for k, v := range sprig.TxtFuncMap() {
		funcMap[k] = v
	}

	funcMap["include"] = func(name string, data any) (string, error) {
		buf := bytes.NewBuffer(nil)
		if err := t.ExecuteTemplate(buf, name, data); err != nil {
			return "", err
		}
		return buf.String(), nil
	}

	_, err := t.Funcs(funcMap).ParseFiles(tFiles...)
	if err != nil {
		return nil, err
	}
	return &kt{Template: t}, nil
}

type kt struct {
	*template.Template
}

func (kt *kt) withValues(v interface{}) ([]byte, error) {
	w := new(bytes.Buffer)
	if err := kt.ExecuteTemplate(w, kt.Name(), v); err != nil {
		return nil, errors.NewEf(err, "could not execute template")
	}
	return w.Bytes(), nil
}

func Parse(f templateFile, values interface{}) ([]byte, error) {
	t, err := useTemplate(f)
	if err != nil {
		return nil, err
	}
	return t.withValues(values)
}

func ParseObject(f templateFile, values interface{}) (client.Object, error) {
	t, err := useTemplate(f)
	if err != nil {
		return nil, err
	}
	b, err := t.withValues(values)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	if err := yaml.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return &unstructured.Unstructured{Object: m}, nil
}

var templateDir = filepath.Join(os.Getenv("PWD"), "pkg/infraClient/templates")

var helperFiles []string

func init() {
	if v, ok := os.LookupEnv("TEMPLATES_DIR"); ok {
		templateDir = v
	}

	helpersDir := filepath.Join(templateDir, "_helpers")
	dir, err := os.ReadDir(helpersDir)
	if err != nil {
		fmt.Printf("ERR listing templateDir.... %+v\n", err)
		panic(err)
	}
	for _, entry := range dir {
		if !entry.IsDir() {
			helperFiles = append(helperFiles, filepath.Join(helpersDir, entry.Name()))
		}
	}
}

type templateFile string

func (tf templateFile) Path() string {
	return filepath.Join(templateDir, string(tf))
}

const (
	WorkerConfig        templateFile = "./worker.tmpl.yml"
	ControlePlaneConfig templateFile = "./control.tmpl.yml"
	TalosConfig         templateFile = "./talosconfig.yml"
)
