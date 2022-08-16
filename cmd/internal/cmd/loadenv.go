/*
Copyright © 2022 Kloudlite <support@kloudlite.io>

*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"kloudlite.io/cmd/internal/lib/server"
)

func load(envs map[string]string, args []string) error {

	var cmd *exec.Cmd

	if len(args) > 0 {
		argsWithoutProg := args[1:]
		cmd = exec.Command(args[0], argsWithoutProg...)
	} else {
		cmd = exec.Command("printenv")
	}

	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	if len(args) > 0 {
		cmd.Env = os.Environ()
	}

	for k, v := range envs {
		if len(args) == 0 {
			fmt.Printf("%s=%q\n", k, v)
		} else {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	if len(args) == 0 {
		return nil
	}

	return cmd.Run()
}

// loadenvCmd represents the loadenv command
var loadenvCmd = &cobra.Command{
	Use:   "loadenv",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println(args)
		appId := TriggerSelectApp()
		app, err := server.GetApp(appId)

		if err != nil {
			fmt.Println(err)
			return
		}

		envs := map[string]string{}

		for _, v := range app.Containers {
			if v.Name != "main" {
				continue
			}
			for _, e := range v.EnvVars {
				if e.Value.Type == "literal" {
					envs[e.Key] = e.Value.Value
				}
			}
		}

		if len(args) == 0 {
			load(envs, []string{})
		} else {
			load(envs, args)
		}

		// fmt.Println(app)
	},
}

func init() {
	rootCmd.AddCommand(loadenvCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// loadenvCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// loadenvCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
