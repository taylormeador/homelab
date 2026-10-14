package cmd

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	pb "github.com/taylormeador/homelab/dexter/internal/gen/dexter/v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

var client pb.DexterClient
var conn *grpc.ClientConn

var rootCmd = &cobra.Command{
	Use:   "dexter",
	Short: "CLI for the homelab platform",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		apiKey := loadAPIKey()
		addr := "localhost:50051"

		var err error
		conn, err = grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("failed to connect: %v", err)
		}

		client = pb.NewDexterClient(conn)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", apiKey)

		cmd.SetContext(ctx)

		cobra.OnFinalize(func() {
			cancel()
			conn.Close()
		})
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

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
