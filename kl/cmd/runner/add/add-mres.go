package add

import (
	"fmt"
	"os"

	"github.com/kloudlite/kl/cmd/box/boxpkg/hashctrl"
	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/domain/server"
	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"

	"github.com/spf13/cobra"
)

var mresCmd = &cobra.Command{
	Use:   "mres [name]",
	Short: "add managed resource references to your kl-config",
	Long: `
This command will add secret entry of managed resource references from current environement to your kl-config file.
`,
	Example: ` 
  kl add mres # add mres secret entry to your kl-config as env var
  kl add  mres [name] # add specific mres secret entry to your kl-config as env var by providing mres name
`,
	Run: func(cmd *cobra.Command, args []string) {

		if err := AddMres(cmd, args); err != nil {
			fn.PrintError(err)
			return
		}

	},
}

func AddMres(cmd *cobra.Command, _ []string) error {
	filePath := fn.ParseKlFile(cmd)
	kt, err := client.GetKlFile(filePath)
	if err != nil {
		return functions.NewE(err)
	}

	//TODO: add changes to the klbox-hash file
	mresName := fn.ParseStringFlag(cmd, "resource")

	mres, err := server.SelectMres([]fn.Option{
		fn.MakeOption("mresName", mresName),
	}...)

	if err != nil {
		return functions.NewE(err)
	}

	mresKey, err := server.SelectMresKey([]fn.Option{
		fn.MakeOption("mresName", mres.Metadata.Name),
	}...)

	if err != nil {
		return functions.NewE(err)
	}

	env, err := client.CurrentEnv()
	if err != nil && kt.DefaultEnv != "" {
		env.Name = kt.DefaultEnv
	}

	currMreses := kt.EnvVars.GetMreses()

	if currMreses == nil {
		currMreses = []client.ResType{
			{
				Name: mres.Metadata.Name,
				Env: []client.ResEnvType{
					{
						Key:    RenameKey(fmt.Sprintf("%s_%s", mres.Metadata.Name, *mresKey)),
						RefKey: *mresKey,
					},
				},
			},
		}
	}

	if currMreses != nil {
		matchedMres := false
		for i, rt := range currMreses {
			if rt.Name == mres.Metadata.Name {
				currMreses[i].Env = append(currMreses[i].Env, client.ResEnvType{
					Key:    RenameKey(fmt.Sprintf("%s_%s", mres.Metadata.Name, *mresKey)),
					RefKey: *mresKey,
				})
				matchedMres = true
				break
			}
		}

		if !matchedMres {
			currMreses = append(currMreses, client.ResType{
				Name: mres.Metadata.Name,
				Env: []client.ResEnvType{
					{
						Key:    RenameKey(fmt.Sprintf("%s_%s", mres.Metadata.Name, *mresKey)),
						RefKey: *mresKey,
					},
				},
			})
		}
	}

	kt.EnvVars.AddResTypes(currMreses, client.Res_mres)
	if err := client.WriteKLFile(*kt); err != nil {
		return functions.NewE(err)
	}

	fn.Log(fmt.Sprintf("added mres %s/%s to your kl-file", mres.Metadata.Name, *mresKey))

	wpath, err := os.Getwd()
	if err != nil {
		return functions.NewE(err)
	}

	if err := hashctrl.SyncBoxHash(wpath); err != nil {
		return functions.NewE(err)
	}

	//if err := server.SyncDevboxJsonFile(); err != nil {
	//	return functions.NewE(err)
	//}
	//
	//if err := client.SyncDevboxShellEnvFile(cmd); err != nil {
	//	return functions.NewE(err)
	//}

	return nil
}

func init() {
	fn.WithKlFile(mresCmd)
}
