package main

import (
	"os"

	"github.com/kloudlite/kloudlite/cli/kl/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
