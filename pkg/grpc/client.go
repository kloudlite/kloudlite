package grpc

import (
	"crypto/tls"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

func Connect(url string) (*grpc.ClientConn, error) {
	for {
		conn, err := grpc.Dial(url, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
		if err == nil {
			return conn, nil
		}
		log.Printf("Failed to connect: %v, retrying...", err)
		time.Sleep(2 * time.Second)
	}
}

func ConnectSecure(url string) (*grpc.ClientConn, error) {
	for {
		conn, err := grpc.Dial(url, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			InsecureSkipVerify: false,
		})), grpc.WithBlock())
		if err == nil {
			return conn, nil
		}
		log.Printf("Failed to connect: %v, retrying...", err)
		time.Sleep(2 * time.Second)
	}
}

// package main
//
// import (
//   "context"
//   "log"
//   "time"
//
//   "google.golang.org/grpc"
//   "google.golang.org/grpc/connectivity"
//   pb "path/to/generated/example"
// )
//
// func main() {
//   address := "localhost:50051"
//
//   for {
//     conn, err := connectWithRetry(address)
//     if err != nil {
//       log.Fatalf("Failed to connect after retries: %v", err)
//     }
//
//     client := pb.NewExampleServiceClient(conn)
//
//     // Use the client to call your gRPC service methods
//     // ...
//
//     // Check the connection state and attempt to reconnect if it is not ready or has been terminated
//     connState := conn.GetState()
//     for connState != connectivity.Ready && connState != connectivity.Shutdown {
//       log.Printf("Connection lost, trying to reconnect...")
//       time.Sleep(2 * time.Second)
//       connState = conn.GetState()
//     }
//
//     conn.Close()
//   }
// }
