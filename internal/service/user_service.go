package service

import (
	"context"

	"github.com/lahaehae/crud_project/internal/pb"
	"github.com/lahaehae/crud_project/internal/repository"
)

// UserService реализует pb.UserServiceServer
type UserService struct {	
	pb.UnimplementedUserServiceServer
	repo repository.UserRepository
}

// NewUserService создаёт новый экземпляр UserService
func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.UserResponse, error) {
	user, err := s.repo.CreateUser(ctx, req.Name, req.Email)
	if err != nil {
		return nil, err
	}
	return &pb.UserResponse{
		Id:    user.Id,
		Name:  user.Name,
		Email: user.Email,
	}, nil
}

func (s *UserService) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.UserResponse, error) {
	user, err := s.repo.GetUser(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &pb.UserResponse{
		Id:    user.Id,
		Name:  user.Name,
		Email: user.Email,
	}, nil
}

func (s *UserService) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UserResponse, error) {
	user, err := s.repo.UpdateUser(ctx, req.Id, req.Name, req.Email)
	if err != nil {
		return nil, err
	}
	return &pb.UserResponse{
		Id:    user.Id,
		Name:  user.Name,
		Email: user.Email,
	}, nil
}

func (s *UserService) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.Empty, error) {
	err := s.repo.DeleteUser(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &pb.Empty{}, nil
}
