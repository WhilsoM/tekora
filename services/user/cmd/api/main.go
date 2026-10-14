package main

import (
	"log"
	"net"
	userpb "tekora/pkg/gen/user"
	"tekora/services/user/internal/adapters/transport"
	"tekora/services/user/internal/core"

	"google.golang.org/grpc"
)

func main() {
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer()

	svc := core.NewUserService()
	grpcTransport := transport.NewServer(svc)

	userpb.RegisterUserServiceServer(grpcServer, grpcTransport)

	log.Println("user grpc server started on :50051")

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal(err)
	}
}
