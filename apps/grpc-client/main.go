package main

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/iam"
)

func addMembership(client iam.IAMClient) {
	res, err := client.AddMembership(context.TODO(), &iam.InAddMembership{
		UserId:       "sample kumar1",
		ResourceType: "account",
		ResourceId:   "res_sample",
		Role:         "account-admin",
	})

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(res)
}

func listMembership(client iam.IAMClient) {
	res, err := client.ListMemberships(context.TODO(), &iam.InListMemberships{
		UserId: "sample kumar",
	})

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(res)
}

func can(client iam.IAMClient) {
	res, err := client.Can(context.TODO(), &iam.InCan{
		UserId:      "sample kumar1",
		ResourceIds: []string{"res_sample"},
		Action:      "delete-account",
	})

	if err != nil {
		fmt.Println("ERR:", err)
	}

	fmt.Println("RES:", res.Status)
}

func main() {
	conn, err := grpc.Dial("localhost:3000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	client := iam.NewIAMClient(conn)

	// addMembership(client)
	// listMembership(client)
	can(client)

	// res, err := client.RemoveMembership(context.TODO(), &iam.InRemoveMembership{
	// 	UserId:     "sample kumar",
	// 	ResourceId: "sdfasdf",
	// })

	// res, err := client.RemoveResource(context.TODO(), &iam.InRemoveResource{
	// 	ResourceId: "sdfasdf",
	// })

	// ping, err := client.Ping(context.Background(), &iam.Message{
	// 	Message: "Hello",
	// })
	// fmt.Println(ping, err)
}
