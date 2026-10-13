package main

import (
	"log"
	"net"
	"os"

	pb "github.com/taylormeador/homelab/dexter/internal/gen/dexter/v1"
	"github.com/taylormeador/homelab/dexter/internal/server"
	"google.golang.org/grpc"
)

func main() {
	url := os.Getenv("GITHUB_REPO_URL")
	token := os.Getenv("GITHUB_ACCESS_TOKEN")
	directory := os.Getenv("LOCAL_REPO_DIR")

	server.CloneRepo(url, directory, token)

	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterDexterServer(grpcServer, &server.Server{})

	log.Println("dexterd listening on :50051")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
