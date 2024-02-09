package handler

import (
	"time"

	"github.com/getlantern/systray"
	ns "github.com/kloudlite/kl/app/handler/name-conts"
	"github.com/kloudlite/kl/domain/server"
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
			if server.CheckDeviceStatus() {
				dev.SetTitle("VPN - Connected")
			} else {
				dev.SetTitle("VPN - Disconnected")
			}

			<-time.After(1 * time.Second)
		}
	}()

	go func() {
		for {
			<-dev.ClickedCh
			h.channel <- ChanelMsg{
				Msg:      "Device clicked",
				Item:     dev,
				ItemName: ns.DeviceBtn,
				Action:   ns.ToggleDevice,
			}
		}
	}()
}
