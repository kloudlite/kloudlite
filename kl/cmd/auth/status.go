package auth

import (
	"github.com/kloudlite/kl/domain/apiclient"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
)

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "get the current user's name and email",
	Run: func(_ *cobra.Command, _ []string) {
		apic, err := apiclient.New()
		if err != nil {
			fn.PrintError(err)
			return
		}
		if u, err := apic.GetCurrentUser(); err != nil {
			fn.PrintError(err)
			return
		} else {
			fn.Printf("You are logged in as %s (%s)\n",
				text.Bold(text.Green(u.Name)),
				text.Blue(u.Email),
			)
			return
		}
	},
}
