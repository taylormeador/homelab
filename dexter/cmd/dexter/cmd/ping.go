package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	pb "github.com/taylormeador/homelab/dexter/internal/gen/dexter/v1"
)

var pingCmd = &cobra.Command{
	Use:   "ping [message]",
	Short: "Ping the API server",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		msg := "hello"
		if len(args) > 0 {
			msg = args[0]
		}

		resp, err := client.Ping(cmd.Context(), &pb.PingRequest{Message: msg})
		if err != nil {
			return fmt.Errorf("ping failed: %v", err)
		}

		fmt.Println(resp.Message)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pingCmd)
}
