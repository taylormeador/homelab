package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	pb "github.com/taylormeador/homelab/dexter/internal/gen/dexter/v1"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List resources on the Dexter cluster",
}

var listMinecraftCmd = &cobra.Command{
	Use:   "minecraft",
	Short: "List Minecraft worlds",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := client.ListMinecraftWorlds(cmd.Context(), &pb.ListMinecraftWorldsRequest{})
		if err != nil {
			return fmt.Errorf("failed to list worlds: %v", err)
		}

		if len(resp.Worlds) == 0 {
			fmt.Println("no worlds found")
			return nil
		}

		for _, world := range resp.Worlds {
			fmt.Println(world)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.AddCommand(listMinecraftCmd)
}
