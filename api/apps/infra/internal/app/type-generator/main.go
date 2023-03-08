package main

import (
	"fmt"
	"os"

	"kloudlite.io/apps/infra/internal/domain/entities"
	gqltypesgenerator "kloudlite.io/pkg/gql-types-generator"
)

func main() {
	types := []interface{}{
		entities.Cluster{},
		entities.MasterNode{},
		entities.CloudProvider{},
		entities.Edge{},
		entities.NodePool{},
	}

	s := gqltypesgenerator.GenerateGraphQLTypes(types, nil)
	err := os.WriteFile("types.graphql", []byte(s), os.ModePerm)
	if err!= nil{
		panic(err)
	}
	fmt.Println(s)
}
