package main

import (
	"context"
	"log"
	"net"
	productpb "tekora/pkg/gen/product"
	"tekora/pkg/postgres"
	"tekora/services/product/internal/adapters/psql"
	"tekora/services/product/internal/adapters/transport"
	"tekora/services/product/internal/config"
	"tekora/services/product/internal/core"
	"time"

	"google.golang.org/grpc"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg := config.MustConfig("services/product/.env")

	port := ":50052"
	listener, err := net.Listen("tcp", port) // #nosec G102 -- intentionally accessible through private Docker network
	if err != nil {
		return err
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
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	defer func() {
		if err := sqlDB.Close(); err != nil {
			log.Printf("failed to close db connection: %v", err)
		}
	}()

	log.Println("PostgreSQL connected")

	grpcServer := grpc.NewServer()

	repo := psql.NewProductRepository(db)
	svc := core.NewProductService(repo)
	grpcTransport := transport.NewProductServer(svc)

	productpb.RegisterProductServiceServer(grpcServer, grpcTransport)

	log.Println("product grpc server started on", port)

	if err := grpcServer.Serve(listener); err != nil {
		return err
	}

	return nil
}
