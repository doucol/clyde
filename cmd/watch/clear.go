package watch

import (
	"github.com/doucol/clyde/internal/flowdata"
	"github.com/spf13/cobra"
)

var WatchClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear all local watch data",
	Long:  `This will delete the local watch data cache, including all local flow data.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return flowdata.Clear()
	},
}
