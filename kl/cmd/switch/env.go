package sw

import (
	"fmt"

	"github.com/kloudlite/kl/domain/server"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"

	"github.com/spf13/cobra"
)

var switchCmd = &cobra.Command{
	Use:   "env",
	Short: "switch to a different environment",
	Long: `Switch Environment
Examples:
  # switch to a different environment
  kl env switch
	`,

	Run: func(cmd *cobra.Command, _ []string) {

		envName := fn.ParseStringFlag(cmd, "envname")
		projectName := fn.ParseStringFlag(cmd, "projectname")
		accountName := fn.ParseStringFlag(cmd, "account")

		projectName, err := server.EnsureProject([]fn.Option{
			fn.MakeOption("projectName", projectName),
			fn.MakeOption("accountName", accountName),
		}...)
		if err != nil {
			fn.PrintError(err)
			return
		}

		projects, err := server.ListProjects()
		if err != nil {
			fn.PrintError(err)
			return
		}

		var projectExists = false
		for _, project := range projects {
			if project.Metadata.Name == projectName {
				projectExists = true
				break
			}
		}
		if !projectExists {
			fn.PrintError(fmt.Errorf("project %s does not exist", projectName))
			return
		}
		env, err := server.SelectEnv(envName, fn.MakeOption("projectName", projectName))
		if err != nil {
			fn.PrintError(err)
			return
		}
		if err := server.UpdateDeviceEnv([]fn.Option{
			fn.MakeOption("envName", env.Metadata.Name),
			fn.MakeOption("projectName", projectName),
		}...); err != nil {
			fn.PrintError(err)
			return
		}

		fn.Log(text.Bold(text.Green("\nSelected Environment and Project:")),
			text.Blue(fmt.Sprintf("\n%s (%s) and %s ", env.DisplayName, env.Metadata.Name, projectName)),
		)
	},
}

func init() {

	switchCmd.Flags().StringP("envname", "e", "", "environment name")
	switchCmd.Flags().StringP("projectname", "p", "", "project name")
	switchCmd.Flags().StringP("account", "a", "", "account name")
}
