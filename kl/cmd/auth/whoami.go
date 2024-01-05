package auth

import (
	"fmt"

	"github.com/kloudlite/kl/domain/server"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
)

var WhoAmICmd = &cobra.Command{
	Use:   "whoami",
	Short: "user logged in to kloudlite",
	Long: `This command provides details of the user logged in to kloudlite.
Example:
  # Login to kloudlite
  kl whoami 

  when you execute the above command it will print the user name associated with the current effective user ID.
	`,
	Run: func(_ *cobra.Command, _ []string) {
		if u, err := server.GetCurrentUser(); err != nil {
			fn.PrintError(err)
			return
		} else {
			fmt.Printf("You are logged in as %s (%s)\n",
				text.Blue(u.Name),
				text.Blue(u.Email),
			)
			return
		}
	},
}
