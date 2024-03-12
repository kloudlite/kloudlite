package completion

import (
	"fmt"
	"github.com/kloudlite/kl/cmd/shell"
	"github.com/kloudlite/kl/constants"
	"github.com/kloudlite/kl/domain/client"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
)

var AutoCompletion = &cobra.Command{
	Use:                   "auto-completion [bash|zsh|fish|powershell]",
	DisableFlagsInUseLine: true,
	Short:                 "Output shell completion code for the specified shell (bash, zsh, fish, or powershell)",
	Long:                  `Output shell completion code for the specified shell (bash, zsh, fish, or powershell)`,
	Example:               `Output shell completion code for the specified shell (bash, zsh, fish, or powershell)`,
	Run: func(cmd *cobra.Command, args []string) {

		s, err := shell.ShellName()
		if err != nil {
			fn.PrintError(err)
			return
		}

		context, err := client.WriteCompletionContext()
		if err != nil {
			fn.PrintError(err)
		}

		if s == constants.BashShell {
			if err := cmd.Root().GenBashCompletionV2(context, true); err != nil {
				fn.PrintError(err)
				return
			}
		}
		if s == constants.ZshShell {
			if err := cmd.Root().GenZshCompletion(context); err != nil {
				fn.PrintError(err)
				return
			}
		}
		if s == constants.FishShell {
			if err := cmd.Root().GenFishCompletion(context, true); err != nil {
				fn.PrintError(err)
				return
			}
		}
		if s == constants.PowerShell {
			if err := cmd.Root().GenPowerShellCompletionWithDesc(context); err != nil {
				fn.PrintError(err)
				return
			}
		}

		completionContext, err := client.GetCompletionContext()
		if err != nil {
			fn.PrintError(err)
			return
		}

		fn.Log(
			fmt.Sprintf("Please add the generated configuration to your"+
				text.Blue(" %s ")+"shell by either sourcing the file"+text.Blue(" %s ")+
				"or appending its contents to your shell's configuration file, ensuring seamless auto-completion for Kloudlite commands.\n\n"+
				text.Bold("For %s:\n")+text.Yellow(func() string {
				if s == "bash" {
					return "echo \"source %s\" >> ~/.bashrc\n"
				}
				if s == "zsh" {
					return "echo \"source %s\" >> ~/.zshrc\n"
				}
				if s == "fish" {
					return "echo \"source %s\" >> ~/.config/fish/config.fish\n"
				}
				if s == "powershell" {
					return "Add-Content $PROFILE \"%s\"\n"
				}
				return "No instructions available for %s\n"
			}()),
				s, completionContext, s, completionContext),
		)
	},
}

func init() {
	AutoCompletion.Aliases = append(AutoCompletion.Aliases, "ac", "comp", "complete", "completions", "auto-completions", "auto-completes", "auto-complete", "auto-completes")
}
