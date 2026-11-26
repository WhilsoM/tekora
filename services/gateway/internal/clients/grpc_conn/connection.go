package grpcconn

import (
	"fmt"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewConnection(addr string) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("create grpc connection to %s: %w", addr, err)
	}

	return conn, nil
}

func CloseConnection(conn *grpc.ClientConn) {
	if err := conn.Close(); err != nil {
		log.Printf("failed to close gRPC connection: %v", err)
	}
}
