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

	userCreateResp, err := c.CreateUser(ctx, &pb.CreateUserRequest{Name: "John", Email: "johnnytest@mail.com"})
	if err != nil {
		log.Fatalf("Failed to create new user: %v", err)
	}
	log.Printf("Created new user: %v", userCreateResp)

	userUpdateResp, err := c.UpdateUser(ctx, &pb.UpdateUserRequest{Id: 1, Name: "Smith", Email: "smith@mail.com"})
	if err != nil {
		log.Fatalf("Failed to update user: %v", err)
	}
	log.Printf("Updated user: %v", userUpdateResp)

	userDeleteResp, err := c.DeleteUser(ctx, &pb.DeleteUserRequest{Id: 2})
	if err != nil {
		log.Fatalf("Failed to delete user: %v", err)
	}

	log.Printf("Deleted user: %v", userDeleteResp)

	userResp2, err := c.GetUser(ctx, &pb.GetUserRequest{Id: 1})
	if err != nil {
		log.Printf("такого юзера больше нет %v", err)
	}

	log.Printf("Received user: %v", userResp2)

}
