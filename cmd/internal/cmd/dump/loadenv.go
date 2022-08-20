package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/hashicorp/go-envparse"
	"github.com/spf13/cobra"
	"kloudlite.io/cmd/internal/cmd/app"
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
and usage of using your command.`,
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println(args)
		appId := app.TriggerSelectApp()

		envsString, err := server.GetEnvs(appId)
		if err != nil {
			fmt.Println(err)
			return
		}

		myReader := strings.NewReader(envsString)

		envs, err := envparse.Parse(myReader)

		if err != nil {
			fmt.Println(err)
			return
		}

		if len(args) == 0 {
			load(envs, []string{})
		} else {
			load(envs, args)
		}

	},
}

func init() {
	// rootCmd.AddCommand(loadenvCmd)
}
