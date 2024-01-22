package handler

import (
	"fmt"

	"github.com/getlantern/systray"
	ns "github.com/kloudlite/kl/app/handler/name-conts"
	"github.com/kloudlite/kl/domain/server"
	fn "github.com/kloudlite/kl/pkg/functions"
)

func (h *handler) ReconEnv() {

	var envTitle, envBtn *systray.MenuItem

	if h.itemMap[ns.EnvTitle] != nil {
		envTitle = h.itemMap[ns.EnvTitle]
	} else {
		envTitle = systray.AddMenuItem("Environment", "")
		h.AddItem(ns.EnvTitle, envTitle)
		envTitle.Disable()
	}

	isFirstTime := false
	if h.itemMap[ns.EnvBtn] != nil {
		envBtn = h.itemMap[ns.EnvBtn]
	} else {
		envBtn = systray.AddMenuItem("...", "")
		h.AddItem(ns.EnvBtn, envBtn)
		isFirstTime = true
	}

	if isFirstTime {
		envBtn.AddSubMenuItem("Projects", "").Disable()

		go func() {
			projects, err := server.ListProjects()
			if err != nil {
				fn.PrintError(err)
				fn.Notify(err)
			}

			for _, p := range projects {
				projectMenu := envBtn.AddSubMenuItem(p.Metadata.Name, "")

				e, err := server.ListEnvs(fn.MakeOption("projectName", p.Metadata.Name))
				if err != nil {
					fn.PrintError(err)
					fn.Notify(err)
				}

				projectMenu.AddSubMenuItem("Environment", "").Disable()
				for _, env := range e {
					projectMenu.AddSubMenuItem(env.Metadata.Name, "")
				}
			}
		}()
	}

	go func() {
		d, err := server.GetDevice()
		if err != nil {
			fn.PrintError(err)
			fn.Notify(err)

			envBtn.SetTitle("...")
			return
		}

		envBtn.SetTitle(fmt.Sprintf("%s/%s", d.ProjectName, d.EnvName))
	}()

}
