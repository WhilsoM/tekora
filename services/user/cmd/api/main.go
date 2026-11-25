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
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg := config.MustConfig("services/user/.env")
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	listener, err := net.Listen("tcp", ":50051") // #nosec G102 -- intentionally accessible through private Docker network
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
		logger.Error("failed to connect postgresql", "error", err)
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		logger.Error("failed to take db for gorm", "error", err)
		return err
	}
	defer func() {
		if err := sqlDB.Close(); err != nil {
			logger.Error("failed to close db connection", "error", err)
		}
	}()

	logger.Info("Postgresql connected")

	grpcServer := grpc.NewServer()

	repo := psql.NewUserRepository(db)
	svc := core.NewUserService(repo)
	grpcTransport := transport.NewUserServer(svc)

	userpb.RegisterUserServiceServer(grpcServer, grpcTransport)

	logger.Info("user grpc server started on :50051")

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
