package packages

import (
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/spinner"

	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search [name]",
	Short: "search for a package",
	Run: func(cmd *cobra.Command, args []string) {
		if err := searchPackages(cmd, args); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func searchPackages(cmd *cobra.Command, args []string) error {

	// parsing name

	name := fn.ParseStringFlag(cmd, "name")
	if name == "" && len(args) > 0 {
		name = args[0]
	}

	showAll := fn.ParseBoolFlag(cmd, "show-all")

	if name == "" {
		return functions.Error("name is required")
	}

	stopSp := spinner.Client.Start(fmt.Sprintf("searching for package %s", name))
	defer stopSp()

	err := ExecCmd(
		fmt.Sprintf("devbox search %s%s", name,
			func() string {
				if showAll {
					return " --show-all"
				}

				return ""
			}(),
		),
		nil, true,
	)

	stopSp()
	if err != nil {
		return functions.NewE(err)
	}

	return nil
}

func ExecCmd(cmdString string, env map[string]string, verbose bool) error {
	r := csv.NewReader(strings.NewReader(cmdString))
	r.Comma = ' '
	cmdArr, err := r.Read()
	if err != nil {
		return functions.NewE(err)
	}
	cmd := exec.Command(cmdArr[0], cmdArr[1:]...)
	if verbose {
		cmd.Stdout = os.Stdout
	}

	// cmd.Env = os.Environ()

	if env == nil {
		env = map[string]string{}
	}

	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	cmd.Stderr = os.Stderr
	// s.Start()
	err = cmd.Run()
	// s.Stop()
	return functions.NewE(err)
}

func init() {
	searchCmd.Flags().StringP("name", "n", "", "name of the package to remove")
	searchCmd.Flags().BoolP("show-all", "a", false, "list all matching packages")
}
