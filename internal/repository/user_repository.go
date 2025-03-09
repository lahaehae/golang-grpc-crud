package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	pb "github.com/lahaehae/crud_project/internal/pb"
	"github.com/lahaehae/crud_project/internal/telemetry"

	//"github.com/lahaehae/crud_project/internal/telemetry"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type UserRepo interface{
	CreateUser(ctx context.Context, name, email string)(*pb.UserResponse, error)
	GetUser(ctx context.Context, id int32) (*pb.UserResponse, error)
	UpdateUser(ctx context.Context, id int32) (*pb.UserResponse, error)
	DeleteUser(ctx context.Context, id int32) error
}

type UserRepository struct {
	db *pgxpool.Pool;
	meter metric.Meter;
	tracer trace.Tracer;
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository{
	return &UserRepository{
		db: db,
		meter: otel.Meter("repository"),
		tracer: otel.Tracer("repository"),
	}
}

func (r *UserRepository) CreateUser(ctx context.Context, name, email string) (*pb.UserResponse, error) {
	ctx, span := r.tracer.Start(ctx, "Repository.CreateUser")
	defer span.End()

	start := time.Now()

	query := "INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id"
	var id int32
	err := r.db.QueryRow(ctx, query, name, email).Scan(&id)
	if err != nil {
		span.RecordError(err)
		telemetry.ErrorCounter.Add(ctx, 1, metric.WithAttributes(
			attribute.Int64("userId: ", int64(id)),
			attribute.String("error.type", fmt.Sprintf("%T", err)),
			attribute.String("error.msg", err.Error()),
			attribute.String("query", query),
			))
		return nil, err
	}
	duration := time.Since(start).Milliseconds()
	span.SetAttributes(
		attribute.Int64("db_query.time_ms", duration),
		attribute.Int64("db_query.user_id", int64(id)),
	)

	if telemetry.RepoLatencyRecorder != nil {
		telemetry.RepoLatencyRecorder.Record(ctx, time.Since(start).Seconds())
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
