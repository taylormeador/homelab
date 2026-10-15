package cmd

import (
	"bufio"
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
		cfg := loadConfig()

		apiKey := envOrConfig("DEXTER_API_KEY", cfg["api-key"])
		if apiKey == "" {
			log.Fatal("api key not configured: set DEXTER_API_KEY or api-key in ~/.config/dexter/config")
		}

		addr := envOrConfig("DEXTER_ADDR", cfg["addr"])
		if addr == "" {
			addr = "localhost:50051"
		}

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

func envOrConfig(envKey, configVal string) string {
	if val := os.Getenv(envKey); val != "" {
		return val
	}
	return configVal
}

func loadConfig() map[string]string {
	cfg := make(map[string]string)

	home, err := os.UserHomeDir()
	if err != nil {
		return cfg
	}

	file, err := os.Open(home + "/.config/dexter/config")
	if err != nil {
		return cfg
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		cfg[strings.TrimSpace(key)] = strings.TrimSpace(val)
	}

	return cfg
}
