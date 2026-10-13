package server

import (
	"context"
	"fmt"
	"log"

	pb "github.com/taylormeador/homelab/dexter/internal/gen/dexter/v1"
)

type Server struct {
	pb.UnimplementedDexterServer
}

func (s *Server) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PingResponse, error) {
	log.Printf("received ping: %s", req.Message)
	return &pb.PingResponse{
		Message: fmt.Sprintf("pong: %s", req.Message),
	}, nil
}

func (s *Server) CreateMinecraftWorld(ctx context.Context, req *pb.CreateMinecraftWorldRequest) (*pb.CreateMinecraftWorldResponse, error) {
	log.Printf("received CreateMinecraftWorld request: %s", req.Message)

	// TODO create yaml file, commit to git, push.

	return &pb.CreateMinecraftWorldResponse{
		Message: fmt.Sprintf("received CreateMinecraftWorld request: %s", req.Message),
	}, nil
}
