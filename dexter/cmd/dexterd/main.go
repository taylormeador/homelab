package main

import (
	"context"
	"fmt"
	"log"
	"net"

	pb "github.com/taylormeador/homelab/dexter/internal/gen/dexter/v1"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedDexterServer
}

func (s *server) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PingResponse, error) {
	log.Printf("received ping: %s", req.Message)
	return &pb.PingResponse{
		Message: fmt.Sprintf("pong: %s", req.Message),
	}, nil
}

func main() {
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterDexterServer(grpcServer, &server{})

	log.Println("dexterd listening on :50051")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
