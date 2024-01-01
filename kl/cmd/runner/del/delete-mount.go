package del

import (
	"fmt"

	"github.com/kloudlite/kl/constants"
	common_util "github.com/kloudlite/kl/lib/common"
	"github.com/kloudlite/kl/lib/server"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/spf13/cobra"
)

var deleteMountCommand = &cobra.Command{
	Use:   "mount",
	Short: "remove one mount from your " + constants.CmdName + "-config",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command.`,
	Run: func(_ *cobra.Command, _ []string) {
		removeConfigMount()
	},
}

func removeConfigMount() {

	klFile, err := server.GetKlFile(nil)

	if err != nil {
		common_util.PrintError(err)
		return
	}

	selectedMount, err := fuzzyfinder.Find(
		klFile.FileMount.Mounts,
		func(i int) string {
			return fmt.Sprintf("%s | %s | %s", klFile.FileMount.Mounts[i].Type, klFile.FileMount.Mounts[i].Path, klFile.FileMount.Mounts[i].Name)
		},
		fuzzyfinder.WithPromptString("Select Config Group >"),
	)

	if err != nil {
		common_util.PrintError(err)
		return
	}

	newMounts := make([]server.FileEntry, 0)
	for i, fe := range klFile.FileMount.Mounts {
		if i == selectedMount {
			continue
		}
		newMounts = append(newMounts, fe)
	}

	klFile.FileMount.Mounts = newMounts

	err = server.WriteKLFile(*klFile)

	if err != nil {
		common_util.PrintError(err)
	}

	fmt.Println("mount removed")
}
