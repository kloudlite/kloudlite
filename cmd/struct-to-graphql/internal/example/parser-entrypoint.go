package main

import (
	"context"
	"flag"
	"fmt"
	"golang.org/x/sync/errgroup"
	"k8s.io/client-go/rest"
	klrrhlzjw8qccqgbl5gspxq6cqh4zbp2 "kloudlite.io/cmd/struct-to-graphql/internal/example/types"
	"kloudlite.io/cmd/struct-to-graphql/pkg/parser"
	"os"
	"path"
	"strings"

	"encoding/json"
	"io"
	"net/http"
	"time"

	apiExtensionsV1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type schemaClient struct {
	kcli *clientset.Clientset
}

func (s schemaClient) GetK8sJsonSchema(name string) (*apiExtensionsV1.JSONSchemaProps, error) {
	ctx, cf := context.WithTimeout(context.TODO(), 2*time.Second)
	defer cf()
	crd, err := s.kcli.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return crd.Spec.Versions[0].Schema.OpenAPIV3Schema, nil
}

func (s schemaClient) GetHttpJsonSchema(url string) (*apiExtensionsV1.JSONSchemaProps, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var m apiExtensionsV1.JSONSchemaProps
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func main() {
	var isDev bool
	var outDir string
	var withPagination string

	flag.BoolVar(&isDev, "dev", false, "--dev")
	flag.StringVar(&outDir, "out-dir", "struct-to-graphql", "--out-dir <dir-name>")
	flag.StringVar(&withPagination, "with-pagination", "", "--with-pagination <type1,type2,type3...>")
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
		"Example": &klrrhlzjw8qccqgbl5gspxq6cqh4zbp2.Example{},
	}

	kcli, err := func() (*clientset.Clientset, error) {
		if isDev {
			return clientset.NewForConfig(&rest.Config{Host: "localhost:8080"})
		}

		cfg, err := rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
		return clientset.NewForConfig(cfg)
	}()
	if err != nil {
		panic(err)
	}

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

	p := parser.NewParser(&schemaClient{kcli: kcli})

	for k, v := range types {
		p.LoadStruct(k, v)
	}

	if err := g.Wait(); err != nil {
		panic(err)
	}

	p.WithPagination(strings.Split(withPagination, ","))

	if err := p.DumpSchema(outDir); err != nil {
		panic(err)
	}
}
