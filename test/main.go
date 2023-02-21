package main

import (
	"fmt"

	"kloudlite.io/pkg/gql-types-generator"

	cmgrV1 "github.com/kloudlite/cluster-operator/apis/cmgr/v1"
	"kloudlite.io/pkg/repos"
)

type MasterNode struct {
	repos.BaseEntity  `bson:",inline" json:",inline"`
	cmgrV1.MasterNode `json:",inline"`
	// Name              string `json:"name,omitempty"`
	// AccountId        repos.ID `json:"accountId,omitempty"`
	// SubDomain        string   `json:"subDomain,omitempty"`
	// KubeConfig       string   `json:"kubeConfig,omitempty"`
}

func main() {
	res := gqltypesgenerator.GenerateGraphQLTypes(MasterNode{}, nil)
	fmt.Println("=========================================================")
	fmt.Println(res)
}
