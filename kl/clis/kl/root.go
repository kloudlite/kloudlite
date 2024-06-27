package kl

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/kloudlite/kl/flags"
	"github.com/kloudlite/kl/pkg/ui/spinner"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use: flags.CliName,
	PersistentPreRun: func(*cobra.Command, []string) {
		if s, ok := os.LookupEnv("KL_DEV"); ok && s == "true" {
			flags.DevMode = "true"
		}

		sigChan := make(chan os.Signal, 1)

		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sigChan

			spinner.Client.Stop()
			os.Exit(1)
		}()
	},

	PostRun: func(*cobra.Command, []string) {
		spinner.Client.Stop()
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Version = flags.Version
}
