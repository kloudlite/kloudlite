package main

import (
	"context"
	"flag"
	"fmt"
	"golang.org/x/sync/errgroup"
	"k8s.io/client-go/rest"
	"kloudlite.io/pkg/k8s"
	"os"
	"path"
	"strings"
)

type arrayFlags []string

func (i *arrayFlags) String() string {
	return "<nothing>"
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}


func main() {
	var isDev bool
	var outputDir string
	var crds arrayFlags

	flag.BoolVar(&isDev, "dev", false, "--dev")
	flag.StringVar(&outputDir, "output", "./", "--outputDir <dir-name>")

  flag.Var(&crds, "crd", "--skip item1 --skip item2")
	flag.Parse()

	kCli, err := func() (k8s.ExtendedK8sClient, error) {
		if isDev {
			return k8s.NewExtendedK8sClient(&rest.Config{Host: "localhost:8080"})
		}
		config, err := rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
		return k8s.NewExtendedK8sClient(config)
	}()

	if err != nil {
		panic(err)
	}

  crdsMap := make(map[string]string, len(crds))

  for i := range   crds {
    sp := strings.Split(crds[i], "=")
    crdsMap[sp[0]] = sp[1]
  }

	// crdsMap := map[string]string{
	// 	"CloudProvider": "cloudproviders.infra.kloudlite.io",
	// 	"Edge":          "edges.infra.kloudlite.io",
	// 	"NodePool":      "nodepools.infra.kloudlite.io",
	// 	"WorkerNode":    "workernodes.infra.kloudlite.io",
	// 	"Cluster":       "clusters.cmgr.kloudlite.io",
	// 	"MasterNode":    "masternodes.cmgr.kloudlite.io",
	// }

	g, ctx := errgroup.WithContext(context.TODO())
	for k := range crdsMap {
		x := k
		g.Go(func() error {
			schema, err := kCli.GetCRDJsonSchema(ctx, crdsMap[x])
			if err != nil {
				return err
			}

			fmt.Println("calling Convert(", x, ")")
			gqlSchema, err := Convert(schema, x)
			if err != nil {
				return err
			}
			return os.WriteFile(path.Join(outputDir, strings.ToLower(x)+".graphqls"), gqlSchema, 0644)
		})
	}

	g.Go(func() error {
		gqlSchema, err := ScalarTypes()
		if err != nil {
			return err
		}
		return os.WriteFile(path.Join(outputDir, "scalars.graphqls"), gqlSchema, 0644)
	})

	if err := g.Wait(); err != nil {
		panic(err)
	}

	fmt.Println("completed ...")
}
