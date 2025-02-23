package watch

import (
	"github.com/doucol/clyde/internal/whisker"
	"github.com/spf13/cobra"
)

var WatchFlowsCmd = &cobra.Command{
	Use:   "flows",
	Short: "Watch calico flows",
	Long:  `Watch live calico flows in near real-time`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return whisker.WatchFlows(cmd.Context())
	},
}
