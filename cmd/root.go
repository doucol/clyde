package cmd

import (
	"errors"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/doucol/clyde/cmd/watch"
	"github.com/doucol/clyde/internal/cmdctx"
	"github.com/doucol/clyde/internal/logger"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog/v2"
)

var (
	kubeConfig, kubeContext, logLevel string
	logStore                          *logger.LogStore
)

var rootCmd = &cobra.Command{
	Use:   "clyde",
	Short: "Project Calico utilities",
	Long:  "clyde\nA collection of Project Calico utilities",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
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
	rootCmd.PersistentFlags().StringVar(&kubeConfig, "kubeconfig", dflt, "absolute path to the kubeconfig file")
	rootCmd.PersistentFlags().StringVar(&kubeContext, "kubecontext", "", "(optional) kubeconfig context to use")
	rootCmd.PersistentFlags().StringVar(&logLevel, "loglevel", "warn", "log level (trace, debug, info, warn, error)")

	// Add all root commands
	rootCmd.AddCommand(watch.WatchCmd, aboutCmd, versionCmd)
}

func Execute() int {
	stopSignal := make(chan os.Signal, 1)
	signal.Notify(stopSignal, os.Interrupt, syscall.SIGTERM)

	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		initLogger()
		// This new context contains our CmdContext, accessible from every cmd func
		//   - the CmdContext also contains the Cancel func which in turn
		//   calls 'cancel()' on the context triggering app shutdown everywhere
		//   through the context <-Done() channel
		cc := cmdctx.NewCmdCtx(kubeConfig, kubeContext)
		ctx := cc.ToContext(cmd.Context())
		cc = cmdctx.CmdCtxFromContext(ctx)
		go func() {
			<-stopSignal
			cc.Cancel()
		}()
		cmd.SetContext(ctx)
	}
	defer closeLogger()
	if err := rootCmd.Execute(); err != nil {
		return -1
	}
	return 0
}

func initLogger() {
	switch logLevel {
	case "trace":
		logrus.SetLevel(logrus.TraceLevel)
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	default:
		panic(errors.New("invalid log level: " + logLevel))
	}

	var err error
	logStore, err = logger.New()
	if err != nil {
		panic(err)
	}
	log.SetOutput(logStore)
	klog.SetOutput(logStore)
	logrus.SetOutput(logStore)
	logrus.Infof("Logger initialized. Log level set to '%s'", logLevel)
}

func closeLogger() {
	if logStore != nil {
		logStore.Close()
		logStore.Dump(os.Stderr)
	}
}
