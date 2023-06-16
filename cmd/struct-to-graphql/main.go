package main

import (
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	_ "github.com/kloudlite/operator/pkg/operator"
	"k8s.io/apimachinery/pkg/util/rand"

	"github.com/spf13/pflag"
	fn "kloudlite.io/pkg/functions"
	"kloudlite.io/pkg/repos"
	t "kloudlite.io/pkg/types"
)

type Project struct {
	repos.BaseEntity `json:",inline"`
	crdsv1.Project   `json:",inline" graphql:"uri=k8s://projects.crds.kloudlite.io"`
	AccountName      string       `json:"accountName"`
	ClusterName      string       `json:"clusterName"`
	SyncStatus       t.SyncStatus `json:"syncStatus"`
}

func main() {
	var structPaths []string
	pflag.StringSliceVar(&structPaths, "struct", nil, "--struct github.com/kloudlite/sample.Main")
	pflag.Parse()

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
	t2 := template.New("code_gen")
	t2.Funcs(sprig.TxtFuncMap())
	t2.Parse(`package main

import (
  {{- range $key, $value := .Imports}}
  {{$value}} {{$key | quote}}
  {{- end }}
  parser "kloudlite.io/cmd/struct-to-graphql/pkg/parser"
  "kloudlite.io/pkg/k8s"
  "k8s.io/client-go/rest"
  "os"
  "golang.org/x/sync/errgroup"
  "context"
  "path"
  "flag"
  "fmt"
)

func main() {
  var outDir string
  var withPagination bool
  flag.StringVar(&outDir, "out-dir", "struct-to-graphql", "--out-dir <dir-name>")
	flag.BoolVar(&withPagination, "with-pagination", false, "--with-pagination")
  flag.Parse()

  stat, err := os.Stat(outDir)
  if err != nil {
    if os.IsNotExist(err) {
      if err := os.MkdirAll(outDir, 0755); err != nil {
        panic(err)
      }
    }
  }

  if stat != nil && !stat.IsDir() {
    panic(fmt.Errorf("out-dir (%s) is not a directory", outDir))
  }

  types := map[string]any{
    {{- range $key, $value :=  .Types}}
    "{{$key}}": {{$value}},
    {{- end }}
  }

  kCli, err := func() (k8s.ExtendedK8sClient, error) {
    return k8s.NewExtendedK8sClient(&rest.Config{Host: "localhost:8080"})
  }()
  if err != nil {
    panic(err)
  }

  g, _ := errgroup.WithContext(context.TODO())

  g.Go(func() error {
    directives, err := parser.Directives()
    if err != nil {
      return err
    }
    return os.WriteFile(path.Join(outDir, "directives.graphqls"), directives, 0644)
  })

  g.Go(func() error {
    scalarTypes, err := parser.ScalarTypes()
    if err != nil {
      panic(err)
    }
    return os.WriteFile(path.Join(outDir, "scalars.graphqls"), scalarTypes, 0644)
  })

  // g.Go(func() error {
  //   k8s_types, err := parser.KloudliteK8sTypes()
  //   if err != nil {
  //     panic(err)
  //   }
  //   return os.WriteFile(path.Join(outDir, "k8s_types.graphqls"), k8s_types, 0644)
  // })

  p := parser.NewParser(kCli)

  for k, v := range types {
    typeName := k
    typeValue := v
    // g.Go(func() error { 
      p.LoadStruct(typeName, typeValue)
    //   return nil
    // })
  }

  if err := g.Wait(); err != nil {
    panic(err)
  }

	if withPagination {
		p.WithPagination()
	}

  if err := p.DumpSchema(outDir); err != nil {
    panic(err)
  }
}
`)

	t2.Execute(os.Stdout, map[string]any{
		"Imports": imports,
		"Types":   values,
	})
}
