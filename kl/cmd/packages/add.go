package packages

import (
	"fmt"
	"os"
	"slices"

	"github.com/kloudlite/kl/cmd/box/boxpkg/hashctrl"
	"github.com/kloudlite/kl/domain/fileclient"
	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"

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
	fc, err := fileclient.New()
	if err != nil {
		return fn.NewE(err)
	}

	klConf, err := fc.GetKlFile("")
	if err != nil {
		return functions.NewE(err)
	}

	name := fn.ParseStringFlag(cmd, "name")
	if name == "" && len(args) > 0 {
		name = args[0]
	}

	if name == "" {
		return functions.Error("name is required")
	}

	p, err := Resolve(cmd.Context(), name)
	if err != nil {
		return functions.NewE(err)
	}

	name = p
	if slices.Contains(klConf.Packages, name) {
		return nil
	}

	klConf.Packages = append(klConf.Packages, name)
	err = fc.WriteKLFile(*klConf)
	if err != nil {
		return functions.NewE(err)
	}

	fn.Println(fmt.Sprintf("Package %s is added successfully", name))

	cwd, err := os.Getwd()
	if err != nil {
		return functions.NewE(err)
	}

	if err := hashctrl.SyncBoxHash(cwd); err != nil {
		return functions.NewE(err)
	}

	return nil
}

func init() {
	addCmd.Flags().StringP("name", "n", "", "name of the package to install")
}
