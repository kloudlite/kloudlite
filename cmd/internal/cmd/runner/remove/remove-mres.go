package remove

import (
	"fmt"

	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/spf13/cobra"
	"kloudlite.io/cmd/internal/common"
	"kloudlite.io/cmd/internal/constants"
	"kloudlite.io/cmd/internal/lib/server"
)

var removeMresCommand = &cobra.Command{
	Use:   "mres",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		removeMreses()
	},
}

func removeMreses() {

	klFile, err := server.GetKlFile(nil)

	if err != nil {
		common.PrintError(err)
		es := "Please run '" + constants.CMD_NAME + " init' if you are not initialized the file already"
		common.PrintError(fmt.Errorf(es))
		return
	}

	if len(klFile.Mres) == 0 {
		es := "No managed resouce added yet in your file"
		common.PrintError(fmt.Errorf(es))
		return
	}

	selectedMresIndex, err := fuzzyfinder.Find(
		klFile.Mres,
		func(i int) string {
			return klFile.Mres[i].Name
		},
		fuzzyfinder.WithPromptString("Select managed service >"),
	)

	if err != nil {
		common.PrintError(err)
	}

	selectedMres := klFile.Mres[selectedMresIndex]

	newMres := []server.ResType{}

	for i, rt := range klFile.Mres {
		if i == selectedMresIndex {
			continue
		}
		newMres = append(newMres, rt)
	}

	klFile.Mres = newMres

	err = server.WriteKLFile(*klFile)
	if err != nil {
		common.PrintError(err)
	}

	fmt.Printf("removed mres %s from %s-file\n", selectedMres.Name, constants.CMD_NAME)

}
