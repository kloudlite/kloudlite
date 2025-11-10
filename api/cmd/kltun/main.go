package main

import (
	"os"

	"github.com/kloudlite/kloudlite/api/cmd/kltun/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
