package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	pb "github.com/taylormeador/homelab/dexter/internal/gen/dexter/v1"
)

var createMinecraftCmd = &cobra.Command{
	Use:   "minecraft [index]",
	Short: "Create a Minecraft server",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var index *int32
		if len(args) > 0 {
			val, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("index must be a number: %s", args[0])
			}
			v := int32(val)
			index = &v
		}

		req := pb.CreateMinecraftWorldRequest{Index: index}
		resp, err := client.CreateMinecraftWorld(cmd.Context(), &req)
		if err != nil {
			return fmt.Errorf("failed to create world: %v", err)
		}

		fmt.Println(resp.Message)
		return nil
	},
}

func init() {
	createCmd.AddCommand(createMinecraftCmd)
}
