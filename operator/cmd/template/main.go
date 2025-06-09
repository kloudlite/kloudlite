package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"sigs.k8s.io/yaml"
)

func NewTemplate(name string) *template.Template {
	t := template.New(name)
	sprigFns := sprig.TxtFuncMap()

	klFuncs := make(template.FuncMap)
	klFuncs["toYAML"] = func(txt any) (string, error) {
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

	klFuncs["pipefail"] = func(failureMsg string, passedValue any) (any, error) {
		if !reflect.ValueOf(passedValue).IsValid() {
			failFn, ok := sprigFns["fail"].(func(string) (string, error))
			if !ok {
				return "", fmt.Errorf("could not convert to fail fn")
			}
			return failFn(failureMsg)
		}
		return passedValue, nil
	}

	return t.Funcs(sprigFns).Funcs(klFuncs)
}

func capitalize(s string) string {
	return strings.ToUpper(s[0:1]) + s[1:]
}

func genValueMap(setArgs []string) map[string]any {
	valueMap := map[string]any{}

	for _, v := range os.Environ() {
		split := strings.Split(v, "=")
		valueMap[split[0]] = split[1]
		valueMap[capitalize(split[0])] = split[1]
	}

	for i := range setArgs {
		split := strings.Split(setArgs[i], "=")
		valueMap[split[0]] = split[1]
		valueMap[capitalize(split[0])] = split[1]
	}

	return valueMap
}

type arrayFlags []string

func (i *arrayFlags) String() string {
	return "<nothing>"
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func main() {
	var templateFile string
	var setArgs arrayFlags
	flag.Var(&setArgs, "set", "--set key=value --set key2=value2")
	flag.StringVar(&templateFile, "t", "", "-t <template-file>")

	// var setArgs []string
	// flag.StringSliceVarP(&setArgs, "set", "s", []string{}, "--set key=value --set key2=value2")
	// flag.StringVarP(&templateFile, "template", "t", "", "-t <template-file>")
	// flag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	flag.Parse()

	if templateFile == "" {
		t := NewTemplate("inline")
		input, err := io.ReadAll(os.Stdin)
		os.Stdin.Close()
		if err != nil {
			panic(err)
		}

		_, err = t.Parse(string(input))
		if err != nil {
			panic(err)
		}
		if err := t.Execute(os.Stdout, genValueMap(setArgs)); err != nil {
			panic(err)
		}

		return
		// panic("no template file specified, bad arguments to program exiting ...")
	}

	baseName := filepath.Base(templateFile)
	t := NewTemplate(baseName)

	if _, err := t.ParseFiles(templateFile); err != nil {
		panic(err)
	}
	if err := t.ExecuteTemplate(os.Stdout, baseName, genValueMap(setArgs)); err != nil {
		panic(err)
	}
}
