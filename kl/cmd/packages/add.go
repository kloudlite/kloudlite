package packages

import (
	"fmt"
	"github.com/kloudlite/kl/cmd/box/boxpkg"
	"github.com/kloudlite/kl/cmd/box/boxpkg/hashctrl"
	"github.com/kloudlite/kl/domain/apiclient"
	"github.com/kloudlite/kl/domain/fileclient"
	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/spinner"
	"os"
	"slices"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "add new package",
	Run: func(cmd *cobra.Command, args []string) {
		fc, err := fileclient.New()
		if err != nil {
			fn.PrintError(err)
			return
		}
		apic, err := apiclient.New()
		if err != nil {
			fn.PrintError(err)
			return
		}
		if err := addPackages(apic, fc, cmd, args); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func addPackages(apic apiclient.ApiClient, fc fileclient.FileClient, cmd *cobra.Command, args []string) error {
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

	name, hashpkg, err := Resolve(cmd.Context(), name)
	if err != nil {
		return functions.NewE(err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return functions.NewE(err)
	}

	c, err := boxpkg.NewClient(cmd, args)
	if err != nil {
		return functions.NewE(err)
	}

	spinner.Client.Pause()
	_, err = fn.Exec(fmt.Sprintf("nix shell nixpkgs/%s --command echo downloaded", hashpkg), nil)
	if err != nil {
		return functions.NewE(err)
	}
	spinner.Client.Resume()
	if slices.Contains(klConf.Packages, name) {
		return nil
	}

	klConf.Packages = append(klConf.Packages, name)
	err = fc.WriteKLFile(*klConf)
	if err != nil {
		return functions.NewE(err)
	}

	if err := hashctrl.SyncBoxHash(apic, fc, cwd); err != nil {
		return functions.NewE(err)
	}

	fn.Println(fmt.Sprintf("Package %s is added successfully", name))

	if err := c.ConfirmBoxRestart(); err != nil {
		return functions.NewE(err)
	}

	return nil
}

func init() {
	addCmd.Flags().StringP("name", "n", "", "name of the package to install")
}
