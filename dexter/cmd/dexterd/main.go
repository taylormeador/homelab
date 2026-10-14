package main

import (
	"log"
	"net"
	"os"

	pb "github.com/taylormeador/homelab/dexter/internal/gen/dexter/v1"
	"github.com/taylormeador/homelab/dexter/internal/server"
	"google.golang.org/grpc"
)

func mustEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("required environment variable %s is not set", key)
	}
	return val
}

func main() {
	url := mustEnv("GITHUB_REPO_URL")
	token := mustEnv("GITHUB_ACCESS_TOKEN")
	directory := mustEnv("LOCAL_REPO_DIR")

	if err := server.CloneRepo(url, directory, token); err != nil {
		log.Fatalf("failed to clone repo: %v", err)
	}

	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(server.LoggingInterceptor))
	pb.RegisterDexterServer(grpcServer, &server.Server{})

	log.Println("dexterd listening on :50051")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
