package remove

import (
	"fmt"

	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/spf13/cobra"
	"kloudlite.io/cmd/internal/common"
	"kloudlite.io/cmd/internal/lib/server"
)

var removeMountCommand = &cobra.Command{
	Use:   "mount",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		removeConfigMount()
	},
}

func removeConfigMount() {

	klFile, err := server.GetKlFile(nil)

	if err != nil {
		common.PrintError(err)
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
		common.PrintError(err)
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
		common.PrintError(err)
	}

	fmt.Println("mount removed")
}
