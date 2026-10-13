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
)

func main() {
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
