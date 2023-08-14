package main

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	_ "github.com/kloudlite/operator/pkg/operator"
	"k8s.io/apimachinery/pkg/util/rand"

	"github.com/spf13/pflag"
	fn "kloudlite.io/pkg/functions"
)

//go:embed parser-entrypoint.go.tmpl
var parserEntrypoint string

func main() {
	var structPaths []string
	pflag.StringSliceVar(&structPaths, "struct", nil, "--struct github.com/kloudlite/sample.Main")
	pflag.Parse()

	if len(structPaths) == 0 {
		panic("no struct paths specified, they should be specified like `--struct github.com/kloudlite/sample.Main`")
	}

	imports := make(map[string]string, len(structPaths))
	values := make(map[string]any, len(structPaths))
	for _, p := range structPaths {
		sp := strings.SplitN(fn.StringReverse(p), ".", 2)
		if len(sp) != 2 {
			panic("invalid struct path")
		}
		sp[0] = fn.StringReverse(sp[0])
		sp[1] = fn.StringReverse(sp[1])

		alias := "kl" + rand.String(30)
		existingAlias, ok := imports[sp[1]]
		if ok {
			alias = existingAlias
		} else {
			imports[sp[1]] = alias
		}

		values[sp[0]] = fmt.Sprintf("&%s.%s{}", alias, sp[0])
	}

	t2 := template.Must(template.New("code_gen").Funcs(sprig.TxtFuncMap()).Parse(parserEntrypoint))

	if err := t2.Execute(os.Stdout, map[string]any{
		"Imports": imports,
		"Types":   values,
	}); err != nil {
		log.Fatal(err)
	}
}
