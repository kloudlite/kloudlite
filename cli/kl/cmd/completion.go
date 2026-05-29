package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for kl.

To load completions:

Bash:
  $ source <(kl completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ kl completion bash > /etc/bash_completion.d/kl
  # macOS:
  $ kl completion bash > $(brew --prefix)/etc/bash_completion.d/kl

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it. You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ kl completion zsh > "${fpath[1]}/_kl"

  # You will need to start a new shell for this setup to take effect.

Fish:
  $ kl completion fish | source

  # To load completions for each session, execute once:
  $ kl completion fish > ~/.config/fish/completions/kl.fish

PowerShell:
  PS> kl completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> kl completion powershell > kl.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			return cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			return cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		}
		return nil
	},
}

func init() {
	RootCmd.AddCommand(completionCmd)
}
