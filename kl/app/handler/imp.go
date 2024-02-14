package handler

import (
	"github.com/getlantern/systray"
	ns "github.com/kloudlite/kl/app/handler/name-conts"
	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/pkg/functions"
)

func (h *handler) ReconAuth() {
	var authBtn *systray.MenuItem
	if h.itemMap[ns.AuthBtn] != nil {
		authBtn = h.itemMap[ns.AuthBtn]
	} else {
		authBtn = systray.AddMenuItem("", "")
		h.AddItem(ns.AuthBtn, authBtn)
	}

	session, err := client.GetAuthSession()
	if err != nil {
		functions.PrintError(err)
		functions.Notify("error:", err)
	}

	if session == "" {
		authBtn.SetTitle("Login")
	} else {
		authBtn.SetTitle("Logout")
	}

	go func() {
		for {
			<-authBtn.ClickedCh

			ses, err := client.GetAuthSession()
			if err != nil {
				functions.PrintError(err)
				functions.Notify("error:", err)
			}

			h.channel <- ChanelMsg{
				Msg:      "Sign out clicked",
				Item:     authBtn,
				ItemName: ns.AuthBtn,
				Action: func() ns.Action {

					if ses == "" {
						return ns.Login
					}

					return ns.Logout
				}(),
			}
		}
	}()
}
