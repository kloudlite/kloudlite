package main

/*
Copyright © 2022 Kloudlite <support@kloudlite.io>
*/

import (
	"fmt"

	"github.com/adrg/xdg"
	"github.com/kloudlite/kl/clis/kl"
	"github.com/kloudlite/kl/clis/kli"
	"github.com/kloudlite/kl/flags"
)

func main() {
	fmt.Println(xdg.CacheHome)
	if flags.CliName == "kl" {
		kl.Execute()
		return
	}

	kli.Execute()
}
