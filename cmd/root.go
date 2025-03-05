package cmd

import (
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/doucol/clyde/cmd/watch"
	"github.com/doucol/clyde/internal/cmdContext"
	"github.com/spf13/cobra"
	"k8s.io/client-go/util/homedir"
)

var rootCmd = &cobra.Command{
	Use:   "clyde",
	Short: "Project Calico utilities",
	Long:  "clyde\nA collection of Project Calico utilities",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var kubeConfig, kubeContext string

func init() {
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
}

func Execute() int {
	stopSignal := make(chan os.Signal, 1)
	signal.Notify(stopSignal, os.Interrupt, syscall.SIGTERM)
	cc := cmdContext.NewCmdContext(kubeConfig, kubeContext)
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		ctx := cc.ToContext(cmd.Context())
		ncc := cmdContext.CmdContextFromContext(ctx)
		go func() {
			<-stopSignal
			ncc.Cancel()
		}()
		cmd.SetContext(ctx)
	}
	if err := rootCmd.Execute(); err != nil {
		return -1
	}
	return 0
}
