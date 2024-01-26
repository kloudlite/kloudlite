package handler

import (
	"fmt"

	"github.com/getlantern/systray"
	ns "github.com/kloudlite/kl/app/handler/name-conts"
	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/domain/server"
	fn "github.com/kloudlite/kl/pkg/functions"
)

func (h *handler) ReconAccount() {
	var accountTitle *systray.MenuItem
	var account *systray.MenuItem

	if h.itemMap[ns.AccountTitle] != nil {
		accountTitle = h.itemMap[ns.AccountTitle]
		h.AddItem(ns.AccountTitle, accountTitle)
	} else {
		accountTitle = systray.AddMenuItem("Account", "")
		accountTitle.Disable()
		h.AddItem(ns.AccountTitle, accountTitle)
	}

	ac, err := client.GetMainCtx()
	if err != nil {
		fn.PrintError(err)
		fn.Notify("Error", err.Error())
	}

	aName := ac.AccountName
	if aName == "" {
		aName = "Select Account"
	}

	isFirstTime := false
	if h.itemMap[ns.AccountBtn] != nil {
		account = h.itemMap[ns.AccountBtn]
		account.SetTitle(aName)
	} else {
		isFirstTime = true
		account = systray.AddMenuItem(aName, "")
		h.AddItem(ns.AccountBtn, account)
	}

	if isFirstTime {
		accounts, err := server.ListAccounts()
		if err != nil {
			fn.PrintError(err)
			fn.Notify("Error", err.Error())
		}

		account.AddSubMenuItem("Switch Accounts", "").Disable()

		for _, a := range accounts {
			cm := account.AddSubMenuItem(a.Metadata.Name, fmt.Sprintf("switch to %s", a.Metadata.Name))

			go func(name string) {
				for {
					select {
					case <-cm.ClickedCh:
						h.channel <- ChanelMsg{
							Msg:      name,
							Item:     cm,
							Action:   ns.SwitchAccount,
							ItemName: ns.AccountItem,
						}
					}
				}

			}(a.Metadata.Name)
		}

	}

	var accountSettings *systray.MenuItem
	if h.itemMap[ns.AccountSettings] != nil {
		accountSettings = h.itemMap[ns.AccountSettings]
	} else {
		accountSettings = systray.AddMenuItem("Account Settings", "")
		h.AddItem(ns.AccountSettings, accountSettings)
	}

	go func() {
		for {
			select {
			case <-accountSettings.ClickedCh:
				h.channel <- ChanelMsg{
					Msg:      "Account Settings clicked",
					Action:   ns.OpenAccountSettings,
					Item:     accountSettings,
					ItemName: ns.AccountSettings,
				}
			}
		}
	}()
}
