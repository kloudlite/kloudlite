package handler

import (
	"time"

	"github.com/getlantern/systray"
	ns "github.com/kloudlite/kl/app/handler/name-conts"
	"github.com/kloudlite/kl/app/icons"
	"github.com/kloudlite/kl/domain/client"
	fn "github.com/kloudlite/kl/pkg/functions"
)

type ChanelMsg struct {
	Msg       string
	Item      *systray.MenuItem
	ItemName  ns.ItemName
	Action    ns.Action
	isLoading bool
}

type Handler interface {
	ReconMeta()
	ReconAccount()
	ReconDevice()
	ReconEnv()
	ReconQuit()
	ReconAuth()
	ReconUser()
	StartListener()
	Channel() chan ChanelMsg

	ItemMap() map[ns.ItemName]*systray.MenuItem
	AddItem(ns.ItemName, *systray.MenuItem)
	DeleteItem(ns.ItemName)
}

func (h *handler) AddItem(name ns.ItemName, item *systray.MenuItem) {
	h.itemMap[name] = item
}

func (h *handler) DeleteItem(name ns.ItemName) {
	delete(h.itemMap, name)
}

type handler struct {
	channel chan ChanelMsg
	itemMap map[ns.ItemName]*systray.MenuItem
}

func NewHandler(channel chan ChanelMsg) Handler {
	return &handler{
		channel: channel,
		itemMap: make(map[ns.ItemName]*systray.MenuItem),
	}
}

func (h *handler) Channel() chan ChanelMsg {
	return h.channel
}

func (h *handler) ItemMap() map[ns.ItemName]*systray.MenuItem {
	return h.itemMap
}

func (h *handler) ReconMeta() {
	systray.SetTemplateIcon(icons.KlIcon, icons.KlIcon)
	systray.SetTooltip("Kloudlite vpn client")

	go func() {
		for {
			b, err := client.IsLoading()
			if err != nil {
				fn.PrintError(err)
				fn.Alert("Error", err.Error())
			}

			if b {
				systray.SetTemplateIcon(icons.LoadingIcon, icons.LoadingIcon)
			} else {
				systray.SetTemplateIcon(icons.KlIcon, icons.KlIcon)
			}

			<-time.After(time.Millisecond * 500)
		}
	}()
}

func (h *handler) ReconQuit() {
	mQuitOrig := systray.AddMenuItem("Quit", "Quit client")
	go func() {
		<-mQuitOrig.ClickedCh
		fn.Println("Requesting quit")
		systray.Quit()
		fn.Println("Finished quitting")
	}()
}
