package packages

import (
	"errors"
	"fmt"
	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/domain/server"
	"strings"

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
	name := fn.ParseStringFlag(cmd, "name")
	if name == "" && len(args) > 0 {
		name = args[0]
	}

	if name == "" {
		return errors.New("name is required")
	}

	klConf, err := client.GetKlFile("")
	splits := strings.Split(name, "@")
	for i, v := range klConf.Packages {
		valSplits := strings.Split(v, "@")
		if splits[0] == valSplits[0] {
			klConf.Packages = append(klConf.Packages[:i], klConf.Packages[i+1:]...)
			break
		}
	}
	err = client.WriteKLFile(*klConf)
	if err != nil {
		return err
	}

	fn.Println(fmt.Sprintf("Package %s is deleted", name))
	if err := server.SyncBoxHash(); err != nil {
		return err
	}
	return nil
}

func init() {
	rmCmd.Flags().StringP("name", "n", "", "name of the package to remove")
	rmCmd.Flags().BoolP("verbose", "v", false, "name of the package to install")
}
