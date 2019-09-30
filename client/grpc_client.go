package client

import (
	"fmt"
	"google.golang.org/grpc"
)

func RpcConnect() *grpc.ClientConn {
	client, err := grpc.Dial("localhost:50052", grpc.WithInsecure())
	if err != nil {
		fmt.Println("Unable to initialize connection to RPC")
		return nil
	}

	return client
}