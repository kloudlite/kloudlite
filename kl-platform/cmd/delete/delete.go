package delete

import (
	"fmt"
	"os"
	"strings"

	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "delete",
	Short: "delete the cluster",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

		delDirs := []string{"/etc/rancher/k3s", "/etc/rancher/k3s", "/var/lib/rancher/k3s"}

		fn.Logf(fmt.Sprint("This will delete directories ", fmt.Sprint("\"", text.Bold(strings.Join(delDirs, ", ")), "\""), " and all its contents,", " are you sure?[y/N] "))

		if !fn.Confirm("y", "N") {
			fn.Log(text.Green("Cancelled"))
			return
		}

		fn.Log()
		for _, v := range delDirs {
			if err := os.RemoveAll(v); err != nil {
				fn.PrintError(err)
				return
			}

			fn.Log("Deleted", text.Bold(v))
		}
	},
}

func init() {
}
