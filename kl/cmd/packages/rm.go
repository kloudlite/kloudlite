package packages

import (
	"fmt"
	domainutil "github.com/kloudlite/kl/domain/util"
	"os"
	"strings"

	"github.com/kloudlite/kl/cmd/box/boxpkg/hashctrl"
	"github.com/kloudlite/kl/domain/fileclient"

	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

var rmCmd = &cobra.Command{
	Use:   "rm",
	Short: "remove installed package",
	Run: func(cmd *cobra.Command, args []string) {
		if err := rmPackages(cmd, args); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func rmPackages(cmd *cobra.Command, args []string) error {
	fc, err := fileclient.New()
	if err != nil {
		return fn.NewE(err)
	}

	name := fn.ParseStringFlag(cmd, "name")
	if name == "" && len(args) > 0 {
		name = args[0]
	}

	if name == "" {
		return fn.Error("name is required")
	}

	klConf, err := fc.GetKlFile("")
	if err != nil {
		return fn.NewE(err)
	}
	splits := strings.Split(name, "@")
	for i, v := range klConf.Packages {
		valSplits := strings.Split(v, "@")
		if splits[0] == valSplits[0] {
			klConf.Packages = append(klConf.Packages[:i], klConf.Packages[i+1:]...)
			break
		}
	}

	if err = fc.WriteKLFile(*klConf); err != nil {
		return fn.NewE(err)
	}

	fn.Println(fmt.Sprintf("Package %s is deleted", name))

	cwd, err := os.Getwd()
	if err != nil {
		return fn.NewE(err)
	}

	if err := hashctrl.SyncBoxHash(cwd); err != nil {
		return fn.NewE(err)
	}
	if err := domainutil.ConfirmBoxRestart(cwd); err != nil {
		return fn.NewE(err)
	}
	return nil
}

func init() {
	rmCmd.Flags().StringP("name", "n", "", "name of the package to remove")
}
