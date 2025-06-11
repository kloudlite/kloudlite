package template

import (
	"bytes"
	"fmt"
	"reflect"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"sigs.k8s.io/yaml"
)

func New(name string) *template.Template {
	t := template.New(name)
	sprigFns := sprig.TxtFuncMap()

	var funcs template.FuncMap = map[string]any{}

	funcs["include"] = func(templateName string, templateData any) (string, error) {
		buf := bytes.NewBuffer(nil)
		if err := t.ExecuteTemplate(buf, templateName, templateData); err != nil {
			return "", err
		}
		return buf.String(), nil
	}

	funcs["toYAML"] = func(txt any) (string, error) {
		a, ok := sprigFns["toPrettyJson"].(func(any) string)
		if !ok {
			panic("could not convert sprig.TxtFuncMap[toPrettyJson] into func(any) string")
		}
		ys, err := yaml.JSONToYAML([]byte(a(txt)))
		if err != nil {
			return "", err
		}
		return string(ys), nil
	}

	funcs["endl"] = func() string {
		return "\n"
	}

	funcs["pipefail"] = func(failureMsg string, passedValue any) (any, error) {
		if !reflect.ValueOf(passedValue).IsValid() {
			failFn, ok := sprigFns["fail"].(func(string) (string, error))
			if !ok {
				return "", fmt.Errorf("could not convert to fail fn")
			}
			return failFn(failureMsg)
		}
		return passedValue, nil
	}

	return t.Funcs(sprigFns).Funcs(funcs)
}
