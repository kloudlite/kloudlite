package runner

import (
	"github.com/spf13/cobra"
	"kloudlite.io/cmd/internal/common"
	"kloudlite.io/cmd/internal/lib/server"
)

// initCmd represents the init command
var InitCommand = &cobra.Command{
	Use:   "init",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		initFile, err := server.GetKlFile(nil)

		if err != nil {
			initFile = &server.KLFileType{
				Version: "v1",
				Name:    "Sample",
				Mres:    make([]server.ResType, 0),
				Configs: make([]server.ResType, 0),
				Secrets: make([]server.ResType, 0),
				Env:     []server.EnvType{{Key: "SAMPLE_ENV", Value: "sample_value"}},
				Ports:   []string{},
				FileMount: server.MountType{
					MountBasePath: "./.mounts",
					Mounts:        make([]server.FileEntry, 0),
				},
			}

		}

		err = server.WriteKLFile(*initFile)

		if err != nil {
			common.PrintError(err)
			return
		}
	},
}
