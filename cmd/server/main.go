package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/jackc/pgx/v5"
	pb "github.com/lahaehae/crud_project/pkg"
	"google.golang.org/grpc"
)

func main() {
	conn, err := pgx.Connect(context.Background(), "postgres://postgres:postgres@localhost:5433/crud_project")
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}

	defer conn.Close(context.Background())

	lis, err := net.Listen("tcp", ":9001")
	if err != nil {
		log.Fatalf("Failed to connect to tcp server at 9001: %v", err)
	}

	grpcServer := grpc.NewServer()

	pb.RegisterUserServiceServer(grpcServer, &server{db: conn})

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

type server struct {
	pb.UnimplementedUserServiceServer
	db *pgx.Conn
}

func (s *server) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.UserResponse, error) {
	var user pb.UserResponse
	err := s.db.QueryRow(ctx, "SELECT id, name, email FROM users WHERE id = $1", req.Id).Scan(&user.Id, &user.Name, &user.Email)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *server) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UserResponse, error) {
	log.Println("Попытка обновить юзера: ", req.Id, req.Name, req.Email)

	result, err := s.db.Exec(ctx, "update users set name=$1, email=$2 where id=$3", req.Name, req.Email, req.Id)
	if err != nil {
		return nil, err
	}
	rowsAffected := result.RowsAffected() // чекаем на обновление
	if rowsAffected == 0 {
		return nil, fmt.Errorf("user with id %v not found ", req.Id)
	}
	user := &pb.UserResponse{
		Id:    req.Id,
		Name:  req.Name,
		Email: req.Email,
	}
	log.Println("Юзер успешно обновлен: ", user.Id, user.Name, user.Email)

	return user, nil
}
