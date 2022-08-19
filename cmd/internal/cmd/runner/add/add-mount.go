package add

import (
	"fmt"

	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/spf13/cobra"
	"kloudlite.io/cmd/internal/common"
	"kloudlite.io/cmd/internal/common/ui/input"
	"kloudlite.io/cmd/internal/constants"
	"kloudlite.io/cmd/internal/lib/server"
)

var addMountCommand = &cobra.Command{
	Use:   "mount",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		klFile, err := server.GetKlFile(nil)

		path := ""

		if err != nil {
			common.PrintError(err)
			es := "Please run '" + constants.CMD_NAME + " init' if you are not initialized the file already"
			common.PrintError(fmt.Errorf(es))
			return
		}

		o := input.Options{
			Placeholder: "/code",
			Prompt:      "Enter Path: ",
			// PromptStyle: style.Styles{},
			// CursorStyle: style.Styles{},
			// Value:       "",
			// CharLimit:   0,
			// Width: 1,
			// Password:    false,
		}

		for {

			path, err = input.InputPrompt(o)
			if err != nil {
				common.PrintError(err)
				return
			}

			if path == "" {
				path = "/code"
			}

			matched := -1
			for i, c := range klFile.FileMount.Mounts {
				if c.Path == path {
					matched = i
					items := []string{"Replace content", "Enter another path"}
					resIndex, err := fuzzyfinder.Find(
						items,
						func(i int) string {
							return items[i]
						},
						fuzzyfinder.WithPromptString("Entered Path already present >"),
					)

					if resIndex == 0 {
						matched = -1
						break
					}

					if err != nil {
						common.PrintError(err)
						return
					}

				}
			}
			if matched == -1 {
				break
			}

		}

		selectConfigMount(path, *klFile)
	},
}

func selectConfigMount(path string, klFile server.KLFileType) {

	csName := []string{"config", "secret"}
	cOrsIndex, err := fuzzyfinder.Find(
		csName,
		func(i int) string {
			return csName[i]
		},
		fuzzyfinder.WithPromptString("Select Config Group >"),
	)
	if err != nil {
		common.PrintError(err)
		return
	}

	cOrs := csName[cOrsIndex]
	items := []server.ConfigORSecret{}
	if cOrs == "config" {
		configs, e := server.GetConfigs()

		if e != nil {
			common.PrintError(e)
			return
		}
		for _, c := range configs {
			items = append(items, server.ConfigORSecret(c))
		}

	} else {
		secrets, e := server.GetSecrets()

		if e != nil {
			common.PrintError(e)
			return
		}

		for _, c := range secrets {
			items = append(items, server.ConfigORSecret{
				Entries: c.Entries,
				Id:      c.Id,
				Name:    c.Name,
			})
		}

	}

	if len(items) == 0 {
		es := fmt.Sprintf("No %ss created yet on server", cOrs)
		common.PrintError(fmt.Errorf(es))
		return
	}

	selectedItemIndex, err := fuzzyfinder.Find(
		items,
		func(i int) string {
			return items[i].Name
		},
		fuzzyfinder.WithPromptString(fmt.Sprintf("Select %s >", cOrs)),
	)

	if err != nil {
		common.PrintError(err)
	}

	selectedItem := items[selectedItemIndex]

	matchedIndex := -1
	for i, fe := range klFile.FileMount.Mounts {
		if fe.Path == path {
			matchedIndex = i
		}
	}

	if matchedIndex == -1 {
		klFile.FileMount.Mounts = append(klFile.FileMount.Mounts, server.FileEntry{
			Type: cOrs,
			Ref:  selectedItem.Id,
			Path: path,
			Name: selectedItem.Name,
		})
	} else {
		klFile.FileMount.Mounts[matchedIndex] = server.FileEntry{
			Type: cOrs,
			Ref:  selectedItem.Id,
			Path: path,
			Name: selectedItem.Name,
		}
	}

	err = server.WriteKLFile(klFile)

	if err != nil {
		common.PrintError(err)
	}

	fmt.Println("Mount added to config")
}
