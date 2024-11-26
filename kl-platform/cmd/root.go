package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/kloudlite/kl/flags"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/spinner"
	"github.com/kloudlite/kloudlite/kl-platform/cmd/delete"
	i "github.com/kloudlite/kloudlite/kl-platform/cmd/init"
	"github.com/kloudlite/kloudlite/kl-platform/cmd/kubectl"
	"github.com/kloudlite/kloudlite/kl-platform/cmd/start"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "klp",
	Short: "klp is a tool for managing kloudlite platform",
	Long:  `This tool can help you with managing kloudlite platform. It can be used to initialize, start, delete and check the status of kloudlite platform.`,

	PersistentPreRun: func(cmd *cobra.Command, args []string) {

		if s, ok := os.LookupEnv("KL_DEV"); ok && s == "true" {
			flags.DevMode = "true"
		} else if ok && s == "false" {
			flags.DevMode = "false"
		}

		verbose := fn.ParseBoolFlag(cmd, "verbose")
		if verbose {
			spinner.Client.SetVerbose(verbose)
			flags.IsVerbose = verbose
		}

		quiet := fn.ParseBoolFlag(cmd, "quiet")
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
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {

	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})

	rootCmd.AddCommand(i.GetCmd())
	rootCmd.AddCommand(start.Cmd)
	// rootCmd.AddCommand(down.Cmd)
	// rootCmd.AddCommand(status.Cmd)
	// rootCmd.AddCommand(update.Cmd)
	rootCmd.AddCommand(delete.Cmd)

	if _, err := exec.LookPath("k9s"); err == nil {
		rootCmd.AddCommand(kubectl.K9sCmd)
	}

	if _, err := exec.LookPath("kubectl"); err == nil {
		rootCmd.AddCommand(kubectl.KubectlCmd)
	}

	for _, c := range rootCmd.Commands() {
		c.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
		c.PersistentFlags().BoolP("quiet", "q", false, "quiet output")
	}
}
