package add

import (
	"fmt"

	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/spf13/cobra"
	"kloudlite.io/cmd/internal/common"
	"kloudlite.io/cmd/internal/constants"
	"kloudlite.io/cmd/internal/lib/server"
)

var addMresCommand = &cobra.Command{
	Use:   "mres",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		selectMreses()
	},
}

func selectMreses() {

	klFile, err := server.GetKlFile(nil)

	if err != nil {
		common.PrintError(err)
		es := "Please run '" + constants.CMD_NAME + " init' if you are not initialized the file already"
		common.PrintError(fmt.Errorf(es))
		return
	}

	mreses, market, err := server.GetMreses()

	if err != nil {
		common.PrintError(err)
		return
	}

	if len(mreses) == 0 {
		es := "No managed services created yet on server"
		common.PrintError(fmt.Errorf(es))
		return
	}

	selectedMsvcIndex, err := fuzzyfinder.Find(
		mreses,
		func(i int) string {
			return mreses[i].Name
		},
		fuzzyfinder.WithPromptString("Select managed service >"),
	)

	if err != nil {
		common.PrintError(err)
	}

	selectedMsvc := mreses[selectedMsvcIndex]

	if len(selectedMsvc.Resources) == 0 {
		es := fmt.Sprintf("No resources found in %s managed service", selectedMsvc.Name)
		common.PrintError(fmt.Errorf(es))
		return
	}

	selectedMresIndex, err := fuzzyfinder.Find(
		selectedMsvc.Resources,
		func(i int) string {
			return selectedMsvc.Resources[i].Name
		},
		fuzzyfinder.WithPromptString(fmt.Sprintf("Select resource of %s >", selectedMsvc.Name)),
	)

	if err != nil {
		common.PrintError(err)
	}

	selectedMres := selectedMsvc.Resources[selectedMresIndex]

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
		es := "Can't find the environment in selected resource"
		common.PrintError(fmt.Errorf(es))
		return
	}

	matchedMres := -1

	for i, rt := range klFile.Mres {
		if rt.Id == selectedMres.Id {
			matchedMres = i
			break
		}
	}

	if len(outputs) == 0 {
		es := "No environment variables found in the selected managed resource\n"
		common.PrintError(fmt.Errorf(es))
	}

	if matchedMres != -1 {
		klFile.Mres[matchedMres].Env = func() []server.ResEnvType {
			env := []server.ResEnvType{}

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
					Name: func() string {
						for _, ret := range klFile.Mres[matchedMres].Env {
							if ret.RefKey == op.Name {
								return ret.Name
							}
						}
						return op.Label
					}(),
					RefKey: op.Name,
				})
			}

			return env
		}()
	} else {

		klFile.Mres = append(klFile.Mres, server.ResType{
			Id:   selectedMres.Id,
			Name: selectedMres.Name,
			Env: func() []server.ResEnvType {
				env := []server.ResEnvType{}
				for _, op := range outputs {
					env = append(env, server.ResEnvType{
						Key:    op.Name,
						RefKey: op.Name,
						Name:   op.Label,
					})
				}

				return env
			}(),
		})

		err = server.WriteKLFile(*klFile)
		if err != nil {
			common.PrintError(err)
		}
	}

	fmt.Printf("added mres %s/%s to your %s-file\n", selectedMsvc.Name, selectedMres.Name, constants.CMD_NAME)

}
