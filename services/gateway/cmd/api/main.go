package main

import (
	"log"
	"net/http"
	"os"
	userpb "tekora/pkg/gen/user"
	userhandler "tekora/services/gateway/internal/clients/user"

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
		grpc.WithTransportCredentials(
			insecure.NewCredentials(),
		),
	)
	if err != nil {
		log.Fatal("failed to create grpc client: ", err)
	}
	defer conn.Close()

	userClient := userpb.NewUserServiceClient(conn)

	userHandler := userhandler.NewUserHandler(userClient)

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	router.Route("/api", func(r chi.Router) {
		userHandler.InitUserHandlers(r)
	})

	log.Println("Gateway started on :8080")

	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}
}
