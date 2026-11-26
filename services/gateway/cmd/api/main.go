package main

import (
	"fmt"
	"log"
	"net/http"
	productpb "tekora/pkg/gen/product"
	userpb "tekora/pkg/gen/user"
	grpcconn "tekora/services/gateway/internal/clients/grpc_conn"
	"tekora/services/gateway/internal/clients/product"
	"tekora/services/gateway/internal/clients/user"
	"tekora/services/gateway/internal/config"
	"tekora/services/gateway/internal/myjwt"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg := config.MustConfig("services/gateway/.env")

	tokenManager, err := myjwt.NewManager(
		cfg.AccessSecretKey,
		cfg.RefreshSecretKey,
	)
	if err != nil {
		return fmt.Errorf("initialize token manager: %w", err)
	}

	userConn, err := grpcconn.NewConnection(cfg.UserGRPCAddr)
	if err != nil {
		return fmt.Errorf("connect to user service: %w", err)
	}
	defer grpcconn.CloseConnection(userConn)

	productConn, err := grpcconn.NewConnection(cfg.ProductGRPCAddr)
	if err != nil {
		return fmt.Errorf("connect to product service: %w", err)
	}
	defer grpcconn.CloseConnection(productConn)

	// User service
	userServiceClient := userpb.NewUserServiceClient(userConn)
	userHandler := user.NewUserClient(userServiceClient, tokenManager)

	// Product service
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

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Println("Gateway started on", cfg.HTTPAddr)

	return server.ListenAndServe()
}
