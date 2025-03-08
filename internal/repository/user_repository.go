package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	pb "github.com/lahaehae/crud_project/internal/pb"
)

type UserRepo interface{
	CreateUser(ctx context.Context, name, email string)(*pb.UserResponse, error)
	GetUser(ctx context.Context, id int32) (*pb.UserResponse, error)
	UpdateUser(ctx context.Context, id int32) (*pb.UserResponse, error)
	DeleteUser(ctx context.Context, id int32) error
}

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository{
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(ctx context.Context, name, email string) (*pb.UserResponse, error) {
	var id int32
	err := r.db.QueryRow(ctx, "INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id", name, email).Scan(&id)
	if err != nil {
		return nil, err
	}
	return &pb.UserResponse{Id: id, Name: name, Email: email}, nil
}

func (r *UserRepository) GetUser(ctx context.Context, id int32) (*pb.UserResponse, error) {
	var user pb.UserResponse
	err := r.db.QueryRow(ctx, "SELECT id, name, email FROM users WHERE id = $1", id).Scan(&user.Id, &user.Name, &user.Email)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) UpdateUser(ctx context.Context, id int32, name, email string) (*pb.UserResponse, error) {
	_, err := r.db.Exec(ctx, "UPDATE users SET name = $1, email = $2 WHERE id = $3", name, email, id)
	if err != nil {
		return nil, err
	}
	return &pb.UserResponse{Id: id, Name: name, Email: email}, nil
}

func (r *UserRepository) DeleteUser(ctx context.Context, id int32) error {
	_, err := r.db.Exec(ctx, "DELETE FROM users WHERE id = $1", id)
	return err
}
