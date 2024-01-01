package add

import (
	"fmt"

	"github.com/kloudlite/kl/constants"
	common_util "github.com/kloudlite/kl/lib/common"
	"github.com/kloudlite/kl/lib/server"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/spf13/cobra"
)

var addMresCommand = &cobra.Command{
	Use:   "mres",
	Short: "add mres to your " + constants.CmdName + "-config file by selection from the all the mres available selected project",
	Long: `Add env from managed resource

Using this command you are able to add a environment from the managed resource present on your project
Examples:
  # add managed resource by selecting one
  kl add mres

  # add managed resource providing resourceid and serviceid 
  kl add mres --resource=<resourceId> --service=<serviceId>
`,
	Run: func(cmd *cobra.Command, _ []string) {
		err := selectAndAddMres(cmd)
		if err != nil {
			common_util.PrintError(err)
			return
		}
	},
}

func selectAndAddMres(cmd *cobra.Command) error {
	resource := cmd.Flag("resource").Value.String()
	service := cmd.Flag("service").Value.String()

	klFile, err := server.GetKlFile(nil)

	if err != nil {
		common_util.PrintError(err)
		es := "please run '" + constants.CmdName + " init' if you are not initialized the file already"
		return fmt.Errorf(es)
	}

	mreses, market, err := server.GetMreses()

	if err != nil {
		return err
	}

	if len(mreses) == 0 {
		return fmt.Errorf("no managed services created yet on server")
	}

	selectedMsvc := &server.Mres{}

	if service != "" {
		for _, m := range mreses {
			if m.Name == service {
				selectedMsvc = m
				break
			}
		}
		return fmt.Errorf("no managed service found with the provided name")
	} else {

		selectedMsvcIndex, e := fuzzyfinder.Find(
			mreses,
			func(i int) string {
				return mreses[i].Name
			},
			fuzzyfinder.WithPromptString("Select managed service >"),
		)

		if e != nil {
			return e
		}
		selectedMsvc = mreses[selectedMsvcIndex]
	}

	if len(selectedMsvc.Resources) == 0 {
		return fmt.Errorf("no resources found in %s managed service", selectedMsvc.Name)
	}

	selectedMres := server.ResourceType{}
	if resource != "" {

		for _, rt := range selectedMsvc.Resources {
			if rt.Name == resource {
				selectedMres = rt
				break
			}
		}

		return fmt.Errorf("no managed resource found with provided resource name")

	} else {

		selectedMresIndex, e := fuzzyfinder.Find(
			selectedMsvc.Resources,
			func(i int) string {
				return selectedMsvc.Resources[i].Name
			},
			fuzzyfinder.WithPromptString(fmt.Sprintf("Select resource of %s >", selectedMsvc.Name)),
		)

		if e != nil {
			return e
		}

		selectedMres = selectedMsvc.Resources[selectedMresIndex]

	}

	var outputs server.Outputs

	for _, mc := range market {
		for _, mmi := range mc.List {
			if !mmi.Active {
				continue
			}
			if mmi.Name == selectedMsvc.Source {
				for _, v := range mmi.Resources {
					if v.Name == selectedMres.ResourceType {
						outputs = v.Outputs
					}
				}
				break
			}
		}
	}

	if outputs == nil {
		return fmt.Errorf("can't find the environment in selected resource")
	}

	matchedMres := -1

	for i, rt := range klFile.Mres {
		if fmt.Sprintf("%s/%s", selectedMsvc.Name, rt.Name) == fmt.Sprintf("%s/%s", selectedMsvc.Name, selectedMres.Name) {
			matchedMres = i
			break
		}
	}

	if len(outputs) == 0 {
		return fmt.Errorf("no environment variables found in the selected managed resource")
	}

	if matchedMres != -1 {
		klFile.Mres[matchedMres].Env = func() []server.ResEnvType {
			env := make([]server.ResEnvType, 0)

			for _, op := range outputs {
				env = append(env, server.ResEnvType{
					Key: func() string {
						for _, ret := range klFile.Mres[matchedMres].Env {
							if ret.RefKey == op.Name {
								return ret.Key
							}
						}
						return op.Name
					}(),
					Name: func() *string {
						for _, ret := range klFile.Mres[matchedMres].Env {
							if ret.RefKey == op.Name {
								return ret.Name
							}
						}
						return &op.Label
					}(),
					RefKey: op.Name,
				})
			}

			return env
		}()
	} else {

		klFile.Mres = append(klFile.Mres, server.ResType{
			Name: fmt.Sprintf("%s/%s", selectedMsvc.Name, selectedMres.Name),
			Env: func() []server.ResEnvType {
				env := make([]server.ResEnvType, 0)
				for _, op := range outputs {
					env = append(env, server.ResEnvType{
						Key:    op.Name,
						RefKey: op.Name,
						Name:   &op.Label,
					})
				}

				return env
			}(),
		})

		err = server.WriteKLFile(*klFile)
		if err != nil {
			return err
		}
	}

	fmt.Printf("added mres %s/%s to your %s-file\n", selectedMsvc.Name, selectedMres.Name, constants.CmdName)
	return nil

}

func init() {
	addMresCommand.Flags().StringP("resource", "", "", "managed resource name")
	addMresCommand.Flags().StringP("service", "", "", "managed service name")
}
