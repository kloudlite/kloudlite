package main

/*
Copyright © 2022 Kloudlite <support@kloudlite.io>
*/

import (
	"github.com/kloudlite/kl/clis/kl"
	"github.com/kloudlite/kl/clis/kli"
)

var Cli string = "kl"
var Version = "development"

func main() {
	if Cli == "kl" {
		kl.Execute()
		return
	}

	kli.Execute()
}
