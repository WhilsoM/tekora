package main

import (
	"log"
	"net/http"
	"os"
	userpb "tekora/pkg/gen/user"
	"tekora/services/gateway/internal/clients/user"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	addr := os.Getenv("USER_GRPC_ADDR")
	if addr == "" {
		addr = "localhost:50051"
	}

	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal("failed to create user grpc connection: ", err)
	}
	defer conn.Close()

	userServiceClient := userpb.NewUserServiceClient(conn)

	userHandler := user.NewUserClient(userServiceClient)

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Route("/api", func(r chi.Router) {
		userHandler.InitUserHandlers(r)
	})

	log.Println("Gateway started on :8080")

	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}
}
