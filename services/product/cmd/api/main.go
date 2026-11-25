package main

import (
	"context"
	"log"
	"log/slog"
	"net"
	"os"
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
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	port := ":50052"
	listener, err := net.Listen("tcp", port) // #nosec G102 -- intentionally accessible through private Docker network
	if err != nil {
		logger.Error("failed to listen tcp connection", "error", err)
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
		logger.Error("failed to open postgres connection", "error", err)
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		logger.Error("failed to get sql db for gorm", "error", err)
		return err
	}
	defer func() {
		if err := sqlDB.Close(); err != nil {
			logger.Error("failed to close db connection", "error", err)
		}
	}()

	logger.Info("Postgresql connected")

	grpcServer := grpc.NewServer()

	repo := psql.NewProductRepository(db)
	svc := core.NewProductService(repo)
	grpcTransport := transport.NewProductServer(svc)

	productpb.RegisterProductServiceServer(grpcServer, grpcTransport)

	logger.Info("product grpc server started", "port", port)

	if err := grpcServer.Serve(listener); err != nil {
		return err
	}

	return nil
}
