// Package cmd provides the root command for the Clyde CLI application.
package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/doucol/clyde/internal/cmdctx"
	"github.com/doucol/clyde/internal/flowcache"
	"github.com/doucol/clyde/internal/flowdata"
	"github.com/doucol/clyde/internal/kube"
	"github.com/doucol/clyde/internal/logger"
	"github.com/doucol/clyde/internal/tui"
	"github.com/doucol/clyde/internal/whisker"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog/v2"
)

var (
	kubeConfig, kubeContext, logLevel, logFile string
	logStore                                   *logger.Logger
)

var rootCmd = &cobra.Command{
	Use:           "clyde",
	Short:         "Project Calico utilities",
	Long:          "clyde\nA collection of Project Calico utilities",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := whisker.DefaultConfig()
		cfg.NewUI = func(fds *flowdata.FlowDataStore, fc *flowcache.FlowCache) whisker.FlowUI {
			return tui.NewFlowApp(fds, fc)
		}
		w := whisker.New(cfg)
		return w.WatchFlows(cmd.Context(), nil)
	},
}

func init() {
	dflt := ""
	if home := homedir.HomeDir(); home != "" {
		dflt = filepath.Join(home, ".kube", "config")
	}
	if kcev := os.Getenv("KUBECONFIG"); kcev != "" {
		dflt = kcev
	}

	// Add all global flags
	rootCmd.PersistentFlags().StringVar(&kubeConfig, "kubeconfig", dflt, "Path to the kubeconfig file to use")
	rootCmd.PersistentFlags().StringVar(&kubeContext, "context", "", "The name of the kubeconfig context to use")
	rootCmd.PersistentFlags().StringVar(&logLevel, "loglevel", "warn", "The log level to use (trace, debug, info, warn, error)")
	rootCmd.PersistentFlags().StringVar(&logFile, "logfile", logger.GetDefaultLogFile(), "The log file to use")

	// Add all root commands
	rootCmd.AddCommand(aboutCmd, versionCmd, clearCmd)
}

func Execute() int {
	stopSignal := make(chan os.Signal, 1)
	signal.Notify(stopSignal, os.Interrupt, syscall.SIGTERM)

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if err := initLogger(); err != nil {
			return err
		}
		source := kube.KubeconfigSourceDefault
		if cmd.Flags().Changed("kubeconfig") {
			source = kube.KubeconfigSourceFlag
		} else if os.Getenv("KUBECONFIG") != "" {
			source = kube.KubeconfigSourceEnv
		}
		cc := cmdctx.NewCmdCtx(kubeConfig, source, kubeContext)
		ctx := cc.ToContext(cmd.Context())
		cc = cmdctx.CmdCtxFromContext(ctx)
		go func() {
			<-stopSignal
			cc.Cancel()
		}()
		cmd.SetContext(ctx)
		return nil
	}
	defer closeLogger()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return -1
	}
	return 0
}

func initLogger() error {
	logger.SetLogFile(logFile)
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		return fmt.Errorf("invalid log level %q: %w", logLevel, err)
	}
	logrus.SetLevel(level)

	logStore, err = logger.NewLogger()
	if err != nil {
		return err
	}
	log.SetOutput(logStore)
	klog.SetOutput(logStore)
	logrus.SetOutput(logStore)
	logrus.Infof("Logger initialized. Log level set to '%s'", logLevel)
	return nil
}

func closeLogger() {
	if logStore != nil {
		logStore.Close()
		if err := logStore.Dump(os.Stderr); err != nil {
			fmt.Fprintf(os.Stderr, "clyde: error dumping log: %v\n", err)
		}
	}
}
