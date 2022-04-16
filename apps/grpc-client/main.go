package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/iam"
)

func main() {
	conn, err := grpc.Dial("localhost:3000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	client := iam.NewIAMClient(conn)

	client.AddMembership(context.TODO(), &iam.InAddMembership{
		UserId:       "sample kumar",
		ResourceType: "",
		ResourceId:   "",
		Role:         "",
		Filter:       "",
	})

	ping, err := client.Ping(context.Background(), &iam.Message{
		Message: "Hello",
	})
	fmt.Println(ping, err)
}
