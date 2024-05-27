package del

// import (
// 	"fmt"

// 	"github.com/kloudlite/kl/domain/client"
// 	fn "github.com/kloudlite/kl/pkg/functions"
// 	"github.com/kloudlite/kl/pkg/ui/fzf"

// 	"github.com/spf13/cobra"
// )

// var deleteMresCommand = &cobra.Command{
// 	Use:   "mres",
// 	Short: "A brief description of your command",
// 	Long: `This command help you to delete environment that that is coming from managed resource

// Examples:
//   # remove mres
//   kl del mres
// `,
// 	Run: func(_ *cobra.Command, _ []string) {
// 		err := removeMreses()
// 		if err != nil {
// 			fn.PrintError(err)
// 			return
// 		}
// 	},
// }

// func removeMreses() error {

// 	klFile, err := client.GetKlFile("")

// 	if err != nil {
// 		fn.PrintError(err)
// 		es := "please run 'kl init' if you are not initialized the file already"
// 		return fmt.Errorf(es)
// 	}

// 	if len(klFile.Mres) == 0 {
// 		es := "no managed resouce added yet in your file"
// 		return fmt.Errorf(es)
// 	}

// 	selectedMres, err := fzf.FindOne(
// 		klFile.Mres,
// 		func(item client.ResType) string {
// 			return item.Name
// 		},
// 		fzf.WithPrompt("Select managed service >"),
// 	)

// 	if err != nil {
// 		return err
// 	}

// 	newMres := make([]client.ResType, 0)

// 	for _, rt := range klFile.Mres {
// 		if rt.Name == selectedMres.Name {
// 			continue
// 		}
// 		newMres = append(newMres, rt)
// 	}

// 	klFile.Mres = newMres

// 	err = client.WriteKLFile(*klFile)
// 	if err != nil {
// 		return err
// 	}

// 	fn.Logf("removed mres %s from kl-file\n", selectedMres.Name)

// 	return nil
// }
