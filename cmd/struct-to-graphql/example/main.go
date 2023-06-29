package main

import (
	"os"
	"reflect"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	_ "github.com/kloudlite/operator/pkg/operator"
	"k8s.io/client-go/rest"
	"kloudlite.io/cmd/struct-to-graphql/pkg/parser"

	"kloudlite.io/pkg/k8s"
	"kloudlite.io/pkg/repos"
	// t "kloudlite.io/pkg/types"
)

type Project struct {
	repos.BaseEntity `json:",inline" graphql:"noinput"`
	crdsv1.Project   `json:",inline" graphql:"uri=k8s://projects.crds.kloudlite.io"`
	// SampleOne        int
	// Sample2          int        `json:"sample2,omitempty"`
	// Sample3          int        `json:",omitempty"`
	// Sample4          int        `json:"-"`
	// SampleItem1      SampleType `json:",inline"`
	// SampleItem2      *SampleType
	// SampleItem3      SampleType `json:"sampleItem3"`
	// SampleItems1     []int
	// SampleItems2     []SampleType          `json:"sampleItem2"`
	// SampleItems4     []SampleType          `json:",omitempty"`
	// SampleMap        map[string]SampleType `json:"sampleMap"`
	// AccountName      string                `json:"accountName"`
	// ClusterName      string                `json:"clusterName"`
	// SyncStatus t.SyncStatus `json:"syncStatus"`
	// EnumItem         string                `json:"enumItem" graphql:"enum=One;Two;Three;Four"`
	// SampleExample struct {
	// 	Example1 string
	// 	Example2 string `json:"example2,omitempty"`
	// }
}

// type SampleType struct {
// 	Item1 string
// 	Item2 string `json:"item2,omitempty"`
// }

func main() {
	kCli, err := func() (k8s.ExtendedK8sClient, error) {
		return k8s.NewExtendedK8sClient(&rest.Config{Host: "localhost:8080"})
	}()
	if err != nil {
		panic(err)
	}

	p := parser.NewParser(kCli)
	p.GenerateGraphQLSchema("Project", "Project", reflect.TypeOf(Project{}))
	p.DebugSchema(os.Stdout)
	// if err := p.DumpSchema("struct-to-graphql"); err != nil {
	// 	panic(err)
	// }
}
