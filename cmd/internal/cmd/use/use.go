package use

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "use",
	Short: "select [ account | project ] to use later with all commands",
	Long: `Select account or project for later use
Examples:
  # select account
  kl use account

  # select account with account id
  kl use account <accountId>

  # select project
  kl use project

  # select project with project id
  kl use project <projectId>
`,
}

func init() {
	Cmd.AddCommand(accountsCmd)
	Cmd.AddCommand(projectsCmd)
}
