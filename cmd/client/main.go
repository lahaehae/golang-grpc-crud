package main

import (
	"context"
	"fmt"
	"log"

	//"os/user"
	"sync"
	"time"

	pb "github.com/lahaehae/crud_project/internal/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"go.opentelemetry.io/otel"
	//"go.opentelemetry.io/otel/metric"
)

var (
	tracer 			= otel.Tracer("grpc-client")
	meter           = otel.Meter("grpc-client")
	//requestsCounter metric.Int64Counter
)

// func initMetrics() {
// 	var err error
// 	requestsCounter, err = meter.Int64Counter("grpc_client_requests_total")
// 	if err != nil {
// 		log.Fatalf("Ошибка создания метрики: %v", err)
// 	}
// }

func main() {
	// initMetrics()

	conn, err := grpc.NewClient(":9001", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server")
	}
	defer conn.Close()

	c := pb.NewUserServiceClient(conn)
	
	ctx, span := tracer.Start(context.Background(), "ClientRequest")
	defer span.End()


	ctx, cancel := context.WithTimeout(context.Background(),10 * time.Second)
	defer cancel()


	// requestsCounter.Add(ctx, 1)
	var wg sync.WaitGroup
	numRequests := 1000


	for i:= 0; i < numRequests; i++ {
		wg.Add(1)
		go func(id int){
			defer wg.Done()
			userCreateResp, err := c.CreateUser(ctx, &pb.CreateUserRequest{
				Name: fmt.Sprintf("User%d", id),
				Email: fmt.Sprintf("user%d@mail.com", id),
			})
			if err != nil{
				log.Printf("Failed to create user %d: %v", id, err)
				return
			}
			log.Printf("Created user: %v", userCreateResp)
		}(i)
		//time.Sleep(5 * time.Millisecond)	
	}

	wg.Wait()
	log.Println("All requests completed")
	// userCreateResp, err := c.CreateUser(ctx, &pb.CreateUserRequest{Name: "John", Email: "johnnytest@mail.com"})
	// if err != nil {
	// 	log.Fatalf("Failed to create new user: %v", err)
	// }
	// log.Printf("Created new user: %v", userCreateResp)

	// userResp, err := c.GetUser(ctx, &pb.GetUserRequest{Id: 1})
	// if err != nil {
	// 	log.Fatalf("Failed to get user: %v", err)
	// }

	// log.Printf("Received user: %v", userResp)

	// userUpdateResp, err := c.UpdateUser(ctx, &pb.UpdateUserRequest{Id: 1, Name: "Smith", Email: "smith@mail.com"})
	// if err != nil {
	// 	log.Fatalf("Failed to update user: %v", err)
	// }
	// log.Printf("Updated user: %v", userUpdateResp)

	// userDeleteResp, err := c.DeleteUser(ctx, &pb.DeleteUserRequest{Id: 2})
	// if err != nil {
	// 	log.Fatalf("Failed to delete user: %v", err)
	// }

	// log.Printf("Deleted user: %v", userDeleteResp)

	// userResp2, err := c.GetUser(ctx, &pb.GetUserRequest{Id: 1})
	// if err != nil {
	// 	log.Printf("такого юзера больше нет %v", err)
	// }

	// log.Printf("Received user: %v", userResp2)

}
