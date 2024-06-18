package packages

import (
	"errors"
	"fmt"
	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/domain/server"
	fn "github.com/kloudlite/kl/pkg/functions"
	"slices"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "add new package",
	Run: func(cmd *cobra.Command, args []string) {
		if err := addPackages(cmd, args); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func addPackages(cmd *cobra.Command, args []string) error {
	name := fn.ParseStringFlag(cmd, "name")
	if name == "" && len(args) > 0 {
		name = args[0]
	}
	if name == "" {
		return errors.New("name is required")
	}
	klConf, err := client.GetKlFile("")
	if slices.Contains(klConf.Packages, name) {
		return nil
	}
	klConf.Packages = append(klConf.Packages, name)
	err = client.WriteKLFile(*klConf)
	if err != nil {
		return err
	}
	fn.Println(fmt.Sprintf("Package %s is added successfully", name))
	if err := server.SyncBoxHash(); err != nil {
		return err
	}
	return nil
}

func init() {
	addCmd.Flags().StringP("name", "n", "", "name of the package to install")
	addCmd.Flags().BoolP("verbose", "v", false, "name of the package to install")
}
