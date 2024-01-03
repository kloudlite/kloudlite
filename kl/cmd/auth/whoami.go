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
	Short: "session check for kloudlite",
	Long: `This command let you login to the kloudlite.
Example:
  # Login to kloudlite
  kl whoami 

  when you execute the above command a link will be opened on your browser. 
  visit your browser and approve there to access your account using this cli.
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
