package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	productpb "tekora/pkg/gen/product"
	userpb "tekora/pkg/gen/user"
	"tekora/services/gateway/internal/clients/product"
	"tekora/services/gateway/internal/clients/user"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	addr := os.Getenv("USER_GRPC_ADDR")
	if addr == "" {
		addr = "localhost:50051"
	}

	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("failed to create user grpc connection: %w", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("failed to close gRPC connection: %v", err)
		}
	}()

	userServiceClient := userpb.NewUserServiceClient(conn)

	userHandler := user.NewUserClient(userServiceClient)

	productAddr := os.Getenv("PRODUCT_GRPC_ADDR")
	if productAddr == "" {
		productAddr = "localhost:50052"
	}

	productConn, err := grpc.NewClient(
		productAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("failed to create product grpc connection: %w", err)
	}
	defer func() {
		if err := productConn.Close(); err != nil {
			log.Printf("failed to close gRPC connection: %v", err)
		}
	}()

	productServiceClient := productpb.NewProductServiceClient(productConn)

	productHandler := product.NewProductClient(productServiceClient)

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Route("/api", func(r chi.Router) {
		userHandler.InitUserHandlers(r)
		productHandler.InitProductHandlers(r)
	})

	log.Println("Gateway started on :8080")

	server := &http.Server{
		Addr:              ":8080",
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		return err
	}

	return nil
}
