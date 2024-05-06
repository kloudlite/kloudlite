package main

/*
Copyright © 2022 Kloudlite <support@kloudlite.io>
*/

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
