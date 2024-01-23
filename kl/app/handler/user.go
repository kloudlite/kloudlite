package handler

import (
	"fyne.io/systray"
	ns "github.com/kloudlite/kl/app/handler/name-conts"
	"github.com/kloudlite/kl/domain/server"
)

func (h *handler) ReconUser() {
	var usr *systray.MenuItem

	if h.itemMap[ns.UserBtn] != nil {
		usr = h.itemMap[ns.UserBtn]
	} else {
		usr = systray.AddMenuItem("...", "User")
		h.AddItem(ns.UserBtn, usr)
		usr.Disable()
	}

	go func() {
		u, err := server.GetCurrentUser()
		if err == nil && u != nil {
			usr.SetTitle(u.Name)
		} else {
			usr.SetTitle("...")
		}
	}()
}
