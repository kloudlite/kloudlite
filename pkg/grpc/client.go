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
