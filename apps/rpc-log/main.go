package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/iam"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	client := iam.NewIAMServiceClient(conn)
	ping, err := client.Ping(context.Background(), &iam.Message{
		Message: "Hello",
	})
	fmt.Println(ping, err)
}
