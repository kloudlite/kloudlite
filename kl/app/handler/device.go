package handler

import (
	"github.com/kloudlite/kl/domain/client"
	"time"

	"github.com/getlantern/systray"
	ns "github.com/kloudlite/kl/app/handler/name-conts"
)

func (h *handler) ReconDevice() {
	var dev, toggleConnection *systray.MenuItem

	if h.itemMap[ns.DeviceBtn] != nil {
		dev = h.itemMap[ns.DeviceBtn]
	} else {
		dev = systray.AddMenuItem("", "Kloudlite VPN")
		h.AddItem(ns.DeviceBtn, dev)
	}

	if h.itemMap[ns.VpnConnectionBtn] != nil {
		toggleConnection = h.itemMap[ns.VpnConnectionBtn]
	} else {
		toggleConnection = systray.AddMenuItem("", "Connect/Disconnect to Kloudlite VPN")
		h.AddItem(ns.VpnConnectionBtn, toggleConnection)
	}

	go func() {
		for {
			data, err := client.GetExtraData()
			if err != nil {
				data.VpnConnected = false
			}
			if data.VpnConnected {
				dev.SetTitle("VPN - Connected")
				toggleConnection.SetTitle("Disconnect")
			} else {
				dev.SetTitle("VPN - Disconnected")
				toggleConnection.SetTitle("Connect")
			}
			dev.Disable()

			<-time.After(1 * time.Second)
		}
	}()

	go func() {
		for {
			<-toggleConnection.ClickedCh
			h.channel <- ChanelMsg{
				Msg:      "Connect clicked",
				Item:     toggleConnection,
				ItemName: ns.VpnConnectionBtn,
				Action:   ns.ToggleDevice,
			}
		}
	}()
}
