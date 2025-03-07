package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	pb "github.com/lahaehae/crud_project/pkg"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	//"go.opentelemetry.io/otel/trace"
)

	

var (
	serviceName = semconv.ServiceNameKey.String("grpc-server");
	tracer         = otel.Tracer("grpc-server")
	meter          = otel.Meter("grpc-server") 
	requestsCounter metric.Int64Counter
	latencyRecorder metric.Float64Histogram
	errorCounter 	metric.Int64Counter
)

func initConn() (*grpc.ClientConn, error) {
	// It connects the OpenTelemetry Collector through local gRPC connection.
	// You may replace `localhost:4317` with your endpoint.
	conn, err := grpc.NewClient("localhost:4317",
		// Note the use of insecure transport here. TLS is recommended in production.
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}

	return conn, err
}

// Initializes an OTLP exporter, and configures the corresponding trace provider.
func initTracerProvider(ctx context.Context, res *resource.Resource, conn *grpc.ClientConn) (func(context.Context) error, error) {
	// Set up a trace exporter
	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// Register the trace exporter with a TracerProvider, using a batch
	// span processor to aggregate spans before export.
	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tracerProvider)

	// Set global propagator to tracecontext (the default is no-op).
	otel.SetTextMapPropagator(propagation.TraceContext{})

	// Shutdown will flush any remaining spans and shut down the exporter.
	return tracerProvider.Shutdown, nil
}

// Initializes an OTLP exporter, and configures the corresponding meter provider.
func initMeterProvider(ctx context.Context, res *resource.Resource, conn *grpc.ClientConn) (func(context.Context) error, error) {
	metricExporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics exporter: %w", err)
	}

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(meterProvider)

	return meterProvider.Shutdown, nil
}


func init() {
	meter = otel.Meter("grpc-server")

	var err error
	requestsCounter, err = meter.Int64Counter(
		"count",
		metric.WithDescription("Общее количество gRPC-запросов"),
	)
	if err != nil {
		log.Printf("Ошибка создания счетчика запросов: %v", err)
	}

	latencyRecorder, err = meter.Float64Histogram(
		"latency",
		metric.WithDescription("Время обработки gRPC-запросов"),
	)
	if err != nil {
		log.Printf("Ошибка создания гистограммы задержек: %v", err)
	}

	errorCounter, err = meter.Int64Counter(
		"grpc_server_errors_total",
		metric.WithDescription("Количество ошибок grpc-запросов"),
	)
	if err != nil{
		log.Printf("Ошбика создания счетчика ошибок")
	}
}

func main() {
	log.Printf("Waiting for connection...")
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		attribute.String("service.name", "grpc-service"),
		attribute.String("service.version", "1.0.0"),
		attribute.String("service.instance.id", "instance-123"),
	)
	

	// Подключение к OpenTelemetry Collector
	otelConn, err := initConn()
	if err != nil {
		log.Fatalf("Ошибка подключения к OTEL Collector: %v", err)
	}
	log.Println("Успешное подключение к OTEL Collector")
	defer otelConn.Close()

	shutdownTracer, err := initTracerProvider(ctx, res, otelConn)
	if err != nil {
		log.Fatalf("Ошибка инициализации трассировки: %v", err)
	}
	log.Println("Успешное инициализация трассировки")
	defer shutdownTracer(ctx)

	shutdownMeter, err := initMeterProvider(ctx, res, otelConn)
	if err != nil {
		log.Fatalf("Ошибка инициализации метрик: %v", err)
	}
	log.Println("Успешное инициализация метрик")
	defer shutdownMeter(ctx)

	//"postgres://postgres:postgres@localhost:5433/crud_project?sslmode=disable"
	// connStr := os.Getenv("DATABASE_URL")
	conn, err := pgxpool.New(context.Background(), "postgres://postgres:postgres@localhost:5433/crud_project?sslmode=disable")
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}

	defer conn.Close()

	_, err = conn.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255),
		email VARCHAR(255)
	);`)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	

	lis, err := net.Listen("tcp", ":9001")
	if err != nil {
		log.Fatalf("Failed to connect to tcp server at 9001: %v", err)
	}

	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.MaxConcurrentStreams(1000),	
	)

	reflection.Register(grpcServer)


	pb.RegisterUserServiceServer(grpcServer, &server{db: conn})

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}

}

type server struct {
	pb.UnimplementedUserServiceServer
	db *pgxpool.Pool
}

func (s *server) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.UserResponse, error) {

    ctx, span := tracer.Start(ctx, "GetUser")
	span.SetAttributes(
		attribute.Int64("userId: ", int64(req.Id)),
		attribute.String("operation: ", "database query"),
	)
    defer span.End()

    start := time.Now()

    if requestsCounter != nil {
        requestsCounter.Add(ctx, 1) // Увеличиваем счётчик запросов
    }

    var user pb.UserResponse

	query := "SELECT id, name, email FROM users WHERE id = $1";

    err := s.db.QueryRow(ctx, query, req.Id).Scan(&user.Id, &user.Name, &user.Email)
    if err != nil {
        span.RecordError(err)
		errorCounter.Add(ctx, 1, metric.WithAttributes(
			attribute.Int64("userId: ", int64(req.Id)),
			attribute.String("error.type", fmt.Sprintf("%T", err)),
			attribute.String("error.msg", err.Error()),
			attribute.String("query", query),
		)) 
        return nil, err
    }

    if latencyRecorder != nil {
        latencyRecorder.Record(ctx, time.Since(start).Seconds()) // Записываем задержку
    }

    return &user, nil
}

func (s *server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.UserResponse, error) {
    ctx, span := tracer.Start(ctx, "CreateUser")
	var user pb.UserResponse
	// span.SetAttributes(
	// 	attribute.Int64("user.Id", user.Id),

	// )
    defer span.End()

    start := time.Now()

    if requestsCounter != nil {
        requestsCounter.Add(ctx, 1) // Увеличиваем счётчик
    }

    // log.Println("Попытка инсертнуть юзера: ", req.Name, req.Email)
    
    err := s.db.QueryRow(ctx, "INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id",
        req.Name, req.Email).Scan(&user.Id) // Возвращаем ID нового пользователя

    if err != nil {
        span.RecordError(err)
        return nil, err
    }

    user.Name = req.Name
    user.Email = req.Email
    log.Println("Юзер успешно инсертнут: ", user.Id, user.Name, user.Email)

    if latencyRecorder != nil {
        latencyRecorder.Record(ctx, time.Since(start).Seconds()) // Записываем задержку
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

func (s *server) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.UserResponse, error) {
	var user pb.UserResponse

	err := s.db.QueryRow(ctx, "SELECT id, name, email FROM users where id = $1", req.Id).Scan(&user.Id, &user.Name, &user.Email)
	if err != nil {
		return nil, err
	}

	result, err := s.db.Exec(ctx, "DELETE FROM users where id = $1", req.Id)
	if err != nil {
		return nil, err
	}
	affectedRows := result.RowsAffected()
	if affectedRows == 0 {
		return nil, fmt.Errorf("юзер с таким айди не найден: %v", req.Id)
	}

	return &user, nil
}
