package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	pb "github.com/taylormeador/homelab/dexter/internal/gen/dexter/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func loadAPIKey() string {
	if key := os.Getenv("DEXTER_API_KEY"); key != "" {
		return key
	}

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("could not determine home directory")
	}

	data, err := os.ReadFile(home + "/.config/dexter/api-key")
	if err != nil {
		log.Fatal("DEXTER_API_KEY not set and ~/.config/dexter/api-key not found")
	}

	key := strings.TrimSpace(string(data))
	if key == "" {
		log.Fatal("~/.config/dexter/api-key is empty")
	}

	return key
}

func main() {
	apiKey := loadAPIKey()
	addr := "localhost:50051"

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewDexterClient(conn)

	msg := "hello"
	if len(os.Args) > 1 {
		msg = strings.Join(os.Args[1:], " ")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", apiKey)
	resp, err := client.Ping(ctx, &pb.PingRequest{Message: msg})
	if err != nil {
		log.Fatalf("ping failed: %v", err)
	}

	fmt.Println(resp.Message)

	respTwo, err := client.CreateMinecraftWorld(ctx, &pb.CreateMinecraftWorldRequest{Message: msg})
	if err != nil {
		log.Fatalf("CreateMinecraftWorld failed: %v", err)
	}

	fmt.Println(respTwo.Message)
}
