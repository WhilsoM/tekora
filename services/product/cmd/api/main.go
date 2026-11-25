package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
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

	shutdownCtx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	serveErr := make(chan error, 1)

	go func() {
		serveErr <- grpcServer.Serve(listener)
	}()

	select {
	case err := <-serveErr:
		if err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			return err
		}

	case <-shutdownCtx.Done():
		logger.Info("shutdown signal received")
	}

	logger.Info("graceful shutdown started")

	shutdownDone := make(chan struct{})

	go func() {
		grpcServer.GracefulStop()
		close(shutdownDone)
	}()

	select {
	case <-shutdownDone:
		logger.Info("grpc server stopped gracefully")

	case <-time.After(20 * time.Second):
		logger.Warn("graceful shutdown timeout, forcing stop")

		grpcServer.Stop()
		<-shutdownDone
	}

	return nil
}
