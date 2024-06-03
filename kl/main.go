package main

import (
	"github.com/kloudlite/kl/clis/kl"
	"github.com/kloudlite/kl/clis/kli"
	"github.com/kloudlite/kl/flags"
)

func main() {
	if flags.CliName == "kl" {
		kl.Execute()
		return
	}

	kli.Execute()
}
