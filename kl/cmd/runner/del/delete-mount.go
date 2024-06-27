package del

// import (
// 	"fmt"

// 	"github.com/kloudlite/kl/domain/fileclient"
// 	common_util "github.com/kloudlite/kl/pkg/functions"
// 	"github.com/kloudlite/kl/pkg/ui/fzf"

// 	"github.com/spf13/cobra"
// )

// var deleteMountCommand = &cobra.Command{
// 	Use:   "mount",
// 	Short: "remove one mount from your kl-config",
// 	Long: `A longer description that spans multiple lines and likely contains examples
// and usage of using your command.`,
// 	Run: func(_ *cobra.Command, _ []string) {
// 		removeConfigMount()
// 	},
// }

// func removeConfigMount() {

// 	klFile, err := fileclient.GetKlFile("")

// 	if err != nil {
// 		common_util.PrintError(err)
// 		return
// 	}

// 	selectedMount, err := fzf.FindOne(
// 		klFile.FileMount.Mounts,
// 		func(item fileclient.FileEntry) string {
// 			return fmt.Sprintf("%s | %s | %s", item.Type, item.Path, item.Name)
// 		},
// 		fzf.WithPrompt("Select Config Group >"),
// 	)

// 	if err != nil {
// 		common_util.PrintError(err)
// 		return
// 	}

// 	newMounts := make([]fileclient.FileEntry, 0)
// 	for _, fe := range klFile.FileMount.Mounts {
// 		if fe.Name == selectedMount.Name {
// 			continue
// 		}
// 		newMounts = append(newMounts, fe)
// 	}

// 	klFile.FileMount.Mounts = newMounts

// 	err = fileclient.WriteKLFile(*klFile)

// 	if err != nil {
// 		common_util.PrintError(err)
// 	}

// 	common_util.Log("mount removed")
// }
