package cmd

import (
	"github.com/doucol/clyde/internal/flowdata"
	"github.com/spf13/cobra"
)

var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear all local data",
	Long:  `This will delete all the data that has been captured and cached locally from the cluster`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return flowdata.Clear()
	},
}
