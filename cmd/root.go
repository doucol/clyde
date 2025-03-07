package cmd

import (
	"errors"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/doucol/clyde/cmd/watch"
	"github.com/doucol/clyde/internal/cmdContext"
	"github.com/doucol/clyde/internal/logger"
	log "github.com/sirupsen/logrus"
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

var (
	ls                                *logger.LogStore
	kubeConfig, kubeContext, logLevel string
)

func init() {
	dflt := ""
	if home := homedir.HomeDir(); home != "" {
		dflt = filepath.Join(home, ".kube", "config")
	}
	if kcev := os.Getenv("KUBECONFIG"); kcev != "" {
		dflt = kcev
	}

	// Add all global flags
	rootCmd.PersistentFlags().StringVar(&kubeConfig, "kubeconfig", dflt, "absolute path to the kubeconfig file")
	rootCmd.PersistentFlags().StringVar(&kubeContext, "kubecontext", "", "(optional) kubeconfig context to use")
	rootCmd.PersistentFlags().StringVar(&logLevel, "logLevel", "info", "log level (debug, info, warn, error)")

	// Add all root commands
	rootCmd.AddCommand(watch.WatchCmd, aboutCmd, versionCmd)
}

func Execute() int {
	stopSignal := make(chan os.Signal, 1)
	signal.Notify(stopSignal, os.Interrupt, syscall.SIGTERM)
	cmdctx := cmdContext.NewCmdContext(kubeConfig, kubeContext)
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		// This new context contains our CmdContext, accessible from every cmd func
		//   - the CmdContext also contains the Cancel func which in turn
		//   calls 'cancel()' on the context triggering app shutdown everywhere
		//   through the context <-Done() channel
		newctx := cmdctx.ToContext(cmd.Context())
		cmdctx := cmdContext.CmdContextFromContext(newctx)
		go func() {
			<-stopSignal
			cmdctx.Cancel()
		}()
		cmd.SetContext(newctx)

		switch logLevel {
		case "debug":
			log.SetLevel(log.DebugLevel)
		case "info":
			log.SetLevel(log.InfoLevel)
		case "warn":
			log.SetLevel(log.WarnLevel)
		case "error":
			log.SetLevel(log.ErrorLevel)
		default:
			panic(errors.New("invalid log level"))
		}

		err := logger.Clear()
		if err != nil {
			panic(err)
		}

		ls, err = logger.New()
		if err != nil {
			panic(err)
		}
		log.SetOutput(ls)
	}
	defer func() {
		if ls == nil {
			return
		}
		err := ls.Dump(os.Stderr)
		if err != nil {
			panic(err)
		}
		err = ls.Close()
		if err != nil {
			panic(err)
		}
	}()
	if err := rootCmd.Execute(); err != nil {
		return -1
	}
	return 0
}
