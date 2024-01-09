package add

import (
	"fmt"

	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/domain/server"
	fn "github.com/kloudlite/kl/pkg/functions"

	"github.com/spf13/cobra"
)

var addMresCommand = &cobra.Command{
	Use:   "mres",
	Short: "add mres to your kl-config file by selection from the all the mres available selected project",
	Long: `Add env from managed resource

Using this command you are able to add a environment from the managed resource present on your project
Examples:
  # add managed resource by selecting one
  kl add mres

  # add managed resource providing resourceid and serviceid 
  kl add mres --resource=<resourceId> --service=<serviceId>
`,
	Run: func(cmd *cobra.Command, _ []string) {

		mresName := fn.ParseStringFlag(cmd, "resource")

		mres, err := server.SelectMres([]fn.Option{
			fn.MakeOption("mresName", mresName),
		}...)

		if err != nil {
			fn.PrintError(err)
			return
		}

		mresKey, err := server.SelectMresKey([]fn.Option{
			fn.MakeOption("mresName", mres.Metadata.Name),
		}...)

		if err != nil {
			fn.PrintError(err)
			return
		}

		kt, err := client.GetKlFile(nil)
		if err != nil {
			fn.PrintError(err)
			return
		}

		if kt.Mres == nil {
			kt.Mres = []client.ResType{
				{
					Name: mres.Metadata.Name,
					Env: []client.ResEnvType{
						{
							Key:    fmt.Sprintf("%s_%s", mres.Metadata.Name, *mresKey),
							RefKey: *mresKey,
						},
					},
				},
			}
		}

		if kt.Mres != nil {
			matchedMres := false
			for i, rt := range kt.Mres {
				if rt.Name == mres.Metadata.Name {
					kt.Mres[i].Env = append(kt.Mres[i].Env, client.ResEnvType{
						Key:    fmt.Sprintf("%s_%s", mres.Metadata.Name, *mresKey),
						RefKey: *mresKey,
					})
					matchedMres = true
					break
				}
			}

			if !matchedMres {
				kt.Mres = append(kt.Mres, client.ResType{
					Name: mres.Metadata.Name,
					Env: []client.ResEnvType{
						{
							Key:    fmt.Sprintf("%s_%s", mres.Metadata.Name, *mresKey),
							RefKey: *mresKey,
						},
					},
				})
			}
		}

		if err := client.WriteKLFile(*kt); err != nil {
			fn.PrintError(err)
			return
		}

		fn.Log(fmt.Sprintf("added mres %s/%s to your kl-file", mres.Metadata.Name, *mresKey))
	},
}

func init() {
	addMresCommand.Flags().StringP("resource", "", "", "managed resource name")
}
