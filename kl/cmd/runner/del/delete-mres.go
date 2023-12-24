package del

import (
	"fmt"

	"github.com/kloudlite/kl/constants"
	"github.com/kloudlite/kl/lib/common"
	"github.com/kloudlite/kl/lib/server"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/spf13/cobra"
)

var deleteMresCommand = &cobra.Command{
	Use:   "mres",
	Short: "A brief description of your command",
	Long: `This command help you to delete environment that that is comming from managed resource

Examples:
  # remove mres
  kl del mres
`,
	Run: func(_ *cobra.Command, _ []string) {
		err := removeMreses()
		if err != nil {
			common.PrintError(err)
			return
		}
	},
}

func removeMreses() error {

	klFile, err := server.GetKlFile(nil)

	if err != nil {
		common.PrintError(err)
		es := "please run '" + constants.CmdName + " init' if you are not initialized the file already"
		return fmt.Errorf(es)
	}

	if len(klFile.Mres) == 0 {
		es := "no managed resouce added yet in your file"
		return fmt.Errorf(es)
	}

	selectedMresIndex, err := fuzzyfinder.Find(
		klFile.Mres,
		func(i int) string {
			return klFile.Mres[i].Name
		},
		fuzzyfinder.WithPromptString("Select managed service >"),
	)

	if err != nil {
		return err
	}

	selectedMres := klFile.Mres[selectedMresIndex]

	newMres := make([]server.ResType, 0)

	for i, rt := range klFile.Mres {
		if i == selectedMresIndex {
			continue
		}
		newMres = append(newMres, rt)
	}

	klFile.Mres = newMres

	err = server.WriteKLFile(*klFile)
	if err != nil {
		return err
	}

	fmt.Printf("removed mres %s from %s-file\n", selectedMres.Name, constants.CmdName)

	return nil
}
