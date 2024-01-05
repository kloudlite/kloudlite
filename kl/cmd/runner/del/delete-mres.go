package del

import (
	"fmt"
	"github.com/kloudlite/kl/domain/client"
	common_util "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/fzf"

	"github.com/kloudlite/kl/constants"
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
			common_util.PrintError(err)
			return
		}
	},
}

func removeMreses() error {

	klFile, err := client.GetKlFile(nil)

	if err != nil {
		common_util.PrintError(err)
		es := "please run '" + constants.CmdName + " init' if you are not initialized the file already"
		return fmt.Errorf(es)
	}

	if len(klFile.Mres) == 0 {
		es := "no managed resouce added yet in your file"
		return fmt.Errorf(es)
	}

	selectedMres, err := fzf.FindOne(
		klFile.Mres,
		func(item client.ResType) string {
			return item.Name
		},
		fzf.WithPrompt("Select managed service >"),
	)

	if err != nil {
		return err
	}

	newMres := make([]client.ResType, 0)

	for _, rt := range klFile.Mres {
		if rt.Name == selectedMres.Name {
			continue
		}
		newMres = append(newMres, rt)
	}

	klFile.Mres = newMres

	err = client.WriteKLFile(*klFile)
	if err != nil {
		return err
	}

	fmt.Printf("removed mres %s from %s-file\n", selectedMres.Name, constants.CmdName)

	return nil
}
