package main

import (
	"context"
	"log"
	"time"

	pb "github.com/lahaehae/crud_project/pkg"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.NewClient("localhost:9001", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server")
	}
	defer conn.Close()

	c := pb.NewUserServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	userResp, err := c.GetUser(ctx, &pb.GetUserRequest{Id: 2})
	if err != nil {
		log.Fatalf("Failed to get user: %v", err)
	}

	log.Printf("Received user: %v", userResp)

	userUpdateResp, err := c.UpdateUser(ctx, &pb.UpdateUserRequest{Id: 1, Name: "Smith", Email: "smith@mail.com"})
	if err != nil {
		log.Fatalf("Failed to update user: %v", err)
	}
	log.Printf("Updated user: %v", userUpdateResp)

}
