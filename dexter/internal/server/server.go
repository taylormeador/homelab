package server

import (
	"context"
	"fmt"

	pb "github.com/taylormeador/homelab/dexter/internal/gen/dexter/v1"
	"github.com/taylormeador/homelab/dexter/internal/repo"
)

type Server struct {
	pb.UnimplementedDexterServer
	r *repo.Repo
}

func NewServer(repo *repo.Repo) *Server {
	return &Server{r: repo}
}

func (s *Server) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PingResponse, error) {
	return &pb.PingResponse{
		Message: fmt.Sprintf("pong: %s", req.Message),
	}, nil
}

func (s *Server) CreateMinecraftWorld(ctx context.Context, req *pb.CreateMinecraftWorldRequest) (*pb.CreateMinecraftWorldResponse, error) {
	worldName, err := s.r.CreateMinecraftWorld(req.Index)
	if err != nil {
		return nil, err
	}

	return &pb.CreateMinecraftWorldResponse{
		Message: fmt.Sprintf("created %s", worldName),
	}, nil
}

func (s *Server) DeleteMinecraftWorld(ctx context.Context, req *pb.DeleteMinecraftWorldRequest) (*pb.DeleteMinecraftWorldResponse, error) {
	worldName, err := s.r.DeleteMinecraftWorld(req.Index)
	if err != nil {
		return nil, err
	}

	return &pb.DeleteMinecraftWorldResponse{
		Message: fmt.Sprintf("deleted %s", worldName),
	}, nil
}
