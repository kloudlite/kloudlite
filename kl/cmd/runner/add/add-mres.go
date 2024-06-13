package add

import (
	"fmt"
	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/domain/server"
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
	mresName := fn.ParseStringFlag(cmd, "resource")

	mres, err := server.SelectMres([]fn.Option{
		fn.MakeOption("mresName", mresName),
	}...)
	filePath := fn.ParseKlFile(cmd)

	if err != nil {
		return err
	}

	mresKey, err := server.SelectMresKey([]fn.Option{
		fn.MakeOption("mresName", mres.Metadata.Name),
	}...)

	if err != nil {
		return err
	}

	kt, err := client.GetKlFile(filePath)
	if err != nil {
		return err
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
		return err
	}

	fn.Log(fmt.Sprintf("added mres %s/%s to your kl-file", mres.Metadata.Name, *mresKey))

	if err := server.SyncDevboxJsonFile(); err != nil {
		return err
	}

	if err := client.SyncDevboxShellEnvFile(cmd); err != nil {
		return err
	}

	return nil
}

func init() {
	// mresCmd.Flags().StringP("name", "n", "", "managed resource name")
	fn.WithKlFile(mresCmd)
}
