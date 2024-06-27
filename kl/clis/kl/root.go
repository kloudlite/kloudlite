package kl

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/kloudlite/kl/flags"
	"github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/spinner"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use: flags.CliName,
	PersistentPreRun: func(cmd *cobra.Command, _ []string) {
		if s, ok := os.LookupEnv("KL_DEV"); ok && s == "true" {
			flags.DevMode = "true"
		}

		verbose := functions.ParseBoolFlag(cmd, "verbose")
		if verbose {
			spinner.Client.SetVerbose(verbose)
			flags.IsVerbose = verbose
		}

		quiet := functions.ParseBoolFlag(cmd, "quiet")
		if quiet {
			spinner.Client.SetQuiet(quiet)
			flags.IsQuiet = quiet
		}

		sigChan := make(chan os.Signal, 1)

		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sigChan

			spinner.Client.Stop()
			os.Exit(1)
		}()
	},

	PersistentPostRun: func(*cobra.Command, []string) {
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

	for _, c := range rootCmd.Commands() {
		c.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
		c.PersistentFlags().BoolP("quiet", "q", false, "quiet output")
	}
}
