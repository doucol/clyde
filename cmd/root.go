package cmd

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/doucol/clyde/cmd/watch"
	"github.com/doucol/clyde/internal/cmdContext"
	"github.com/spf13/cobra"
	"k8s.io/client-go/util/homedir"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "clyde",
	Short: "Project Calico utilities",
	Long:  "clyde\nA collection of Project Calico utilities",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var (
	kubeConfig, kubeContext string
	stopSignal              = make(chan os.Signal, 1)
)

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	dflt := ""
	if home := homedir.HomeDir(); home != "" {
		dflt = filepath.Join(home, ".kube", "config")
	}
	if kcev := os.Getenv("KUBECONFIG"); kcev != "" {
		dflt = kcev
	}
	rootCmd.PersistentFlags().StringVar(&kubeConfig, "kubeconfig", dflt, "absolute path to the kubeconfig file")
	rootCmd.PersistentFlags().StringVar(&kubeContext, "kubecontext", "", "(optional) kubeconfig context to use")
	rootCmd.AddCommand(watch.WatchCmd, aboutCmd, versionCmd)
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() int {
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		signal.Notify(stopSignal, os.Interrupt, syscall.SIGTERM)
		ctx, cancel := context.WithCancel(ctx)
		go func() {
			<-stopSignal
			cancel()
		}()
		cmdContext := cmdContext.NewCmdContext(kubeConfig, kubeContext)
		cmd.SetContext(cmdContext.ToContext(ctx))
	}
	if err := rootCmd.Execute(); err != nil {
		return -1
	}
	return 0
}
