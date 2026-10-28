package main

import (
	"context"
	"log"
	"net"
	userpb "tekora/pkg/gen/user"
	"tekora/pkg/postgres"
	"tekora/services/user/internal/adapters/psql"
	"tekora/services/user/internal/adapters/transport"
	"tekora/services/user/internal/config"
	"tekora/services/user/internal/core"
	"time"

	"google.golang.org/grpc"
)

func main() {
	cfg := config.MustConfig("services/user/.env")

	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		8*time.Second,
	)
	defer cancel()

	db, err := postgres.Open(ctx, cfg.DatabaseURL, postgres.PoolConfig{
		MaxOpenConns:    cfg.MaxOpenConns,
		MaxIdleConns:    cfg.MaxIdleConns,
		ConnMaxLifetime: cfg.ConnMaxLifetime,
		ConnMaxIdleTime: cfg.ConnMaxIdleTime,
	})
	if err != nil {
		log.Fatal(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}
	defer sqlDB.Close()

	log.Println("PostgreSQL connected")

	grpcServer := grpc.NewServer()

	repo := psql.NewUserRepository(db)
	svc := core.NewUserService(repo)
	grpcTransport := transport.NewUserServer(svc)

	userpb.RegisterUserServiceServer(grpcServer, grpcTransport)

	log.Println("user grpc server started on :50051")

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal(err)
	}
}
