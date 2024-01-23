package handler

import (
	"time"

	"fyne.io/systray"
	ns "github.com/kloudlite/kl/app/handler/name-conts"
	"github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/wg_vpn/wgc"
)

func (h *handler) ReconDevice() {
	var dev *systray.MenuItem

	if h.itemMap[ns.DeviceBtn] != nil {
		dev = h.itemMap[ns.DeviceBtn]
	} else {
		dev = systray.AddMenuItem("", "Kloudlite VPN")
		h.AddItem(ns.DeviceBtn, dev)
	}

	go func() {

		for {
			wgInterface, err := wgc.Show(&wgc.WgShowOptions{
				Interface: "interfaces",
			})

			if err != nil {
				functions.PrintError(err)
				functions.Notify(err)
			}

			if len(wgInterface) == 0 {
				dev.SetTitle("VPN - Disconnected")
			} else {
				dev.SetTitle("VPN - Connected")
			}

			<-time.After(1 * time.Second)
		}
	}()
}
