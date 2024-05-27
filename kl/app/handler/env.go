package handler

import (
	"github.com/getlantern/systray"
	ns "github.com/kloudlite/kl/app/handler/name-conts"
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

	if h.itemMap[ns.EnvBtn] != nil {
		envBtn = h.itemMap[ns.EnvBtn]
		envBtn.SetTitle("...")
	} else {
		envBtn = systray.AddMenuItem("...", "")
		h.AddItem(ns.EnvBtn, envBtn)
	}
}
