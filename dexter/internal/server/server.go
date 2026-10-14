package server

import (
	"context"
	"fmt"

	pb "github.com/taylormeador/homelab/dexter/internal/gen/dexter/v1"
)

type Server struct {
	pb.UnimplementedDexterServer
}

func (s *Server) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PingResponse, error) {
	return &pb.PingResponse{
		Message: fmt.Sprintf("pong: %s", req.Message),
	}, nil
}

func (s *Server) CreateMinecraftWorld(ctx context.Context, req *pb.CreateMinecraftWorldRequest) (*pb.CreateMinecraftWorldResponse, error) {
	var index int
	if req.Index != nil {
		index = int(req.GetIndex())
	} else {
		var err error
		index, err = getNextWorldIndex()
		if err != nil {
			return nil, err
		}
	}

	if err := createMinecraftYAML(index); err != nil {
		return nil, err
	}

	// TODO commit and push

	return &pb.CreateMinecraftWorldResponse{
		Message: fmt.Sprintf("created minecraft-%d", index),
	}, nil
}
