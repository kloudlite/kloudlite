package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"kloudlite.io/cmd/mocki/internal/parser"
	"log"
	"os"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

func main() {
	var interfaceName string

	var packagePath string
	flag.StringVar(&interfaceName, "interface", "", "--interface <interface>")
	flag.StringVar(&packagePath, "package", "", "--package <package>")
	flag.Parse()

	if interfaceName == "" || packagePath == "" {
		fmt.Println("Invalid values for flags")
		flag.Usage()
		os.Exit(1)
	}

	p := parser.NewParser()

	info, err := p.FindAndParseInterface(packagePath, interfaceName)
	if err != nil {
		log.Fatal(err)
	}

	t := template.New("code_gen")
	t.Funcs(sprig.TxtFuncMap())

	t.Parse(`package {{ .Package }}

import (
  {{- range .Imports }}
  {{.}}
  {{- end }}
)

type {{.InterfaceName}}CallerInfo struct {
	Args []any
}

type {{.StructName}} struct {
  Calls map[string][]{{.InterfaceName}}CallerInfo
  {{- range .MockFunctions }}
  {{ . }}
  {{- end }}
}

func (m *{{.ReceiverStructName}}) registerCall(funcName string, args ...any) {
  if m.Calls == nil {
    m.Calls = map[string][]{{.InterfaceName}}CallerInfo{}
  }
  m.Calls[funcName] = append(m.Calls[funcName], {{.InterfaceName}}CallerInfo{Args: args})
}

{{- "\n" }}
{{- range .Implementations }}
{{ . }}
{{- "\n" }}
{{- end }}

func New{{.StructName}}() *{{.ReceiverStructName}} {
	return &{{.ReceiverStructName}}{}
}
`)

	imports := make([]string, 0, len(info.Imports))
	for _, v := range info.Imports {
		imports = append(imports, fmt.Sprintf("%s %q", v.Alias, v.PackagePath))
	}

	buff := new(bytes.Buffer)
	if err := t.ExecuteTemplate(buff, "code_gen", map[string]any{
		"Package":            "mocks",
		"Implementations":    info.Implementations,
		"StructName":         info.StructName,
		"ReceiverStructName": info.ReceiverStructName,
		"InterfaceName":      interfaceName,
		"MockFunctions":      info.MockFunctions,
		"Imports":            imports,
	}); err != nil {
		log.Fatal(err)
	}

	source, err := format.Source(buff.Bytes())
	if err != nil {
		log.Println("error formatting source:")
		log.Println(buff.String())
		log.Fatal(err)
	}
	fmt.Fprintf(os.Stdout, "%s", source)
	//fmt.Fprintf(os.Stdout, "%s", buff.Bytes())
}
