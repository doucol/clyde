package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	date     = "n/a"
	revision = "dev"
	version  = "dev"
)

var aboutCmd = &cobra.Command{
	Use:   "about",
	Short: "Print version, revision, and date of build",
	Long:  `Print version, revision, and date of build`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version: %s\nRevision: %s\nDate: %s\n", version, revision, date)
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version",
	Long:  `Print version`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version)
	},
}
