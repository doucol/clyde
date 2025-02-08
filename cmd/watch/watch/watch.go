package watch

import (
	"github.com/spf13/cobra"
)

var WatchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch a variety of Calico resources",
	Long:  `Watch a variety of Calico resources`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	WatchCmd.AddCommand(WatchFlowsCmd)
}
