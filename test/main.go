package main

import (
	"context"
	"fmt"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/container_registry"
	rpc "kloudlite.io/pkg/grpc"
)

func main() {
	conn, err := rpc.NewInsecureClient("localhost:50051")
	if err != nil {
		panic(err)
	}
	client := container_registry.NewContainerRegistryClient(conn)
	//account, err := client.CreateProjectForAccount(context.Background(), &container_registry.CreateProjectIn{
	//	AccountName: "test",
	//})
	credentials, err := client.GetSvcCredentials(context.Background(), &container_registry.GetSvcCredentialsIn{
		AccountName: "test",
	})
	fmt.Println(credentials, err)
}
