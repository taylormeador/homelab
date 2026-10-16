package main

import (
	"log"
	"net"
	"os"

	pb "github.com/taylormeador/homelab/dexter/internal/gen/dexter/v1"
	"github.com/taylormeador/homelab/dexter/internal/repo"
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
	adminKey := mustEnv("ADMIN_API_KEY")
	demoKey := mustEnv("DEMO_API_KEY")

	r, err := repo.NewRepo(url, directory, token)
	if err != nil {
		log.Fatalf("failed to init repo: %v", err)
	}

	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(server.LoggingInterceptor, server.NewAuthInterceptor(adminKey, demoKey)))
	pb.RegisterDexterServer(grpcServer, server.NewServer(r))

	log.Println("dexterd listening on :50051")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
