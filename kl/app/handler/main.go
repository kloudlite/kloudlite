package handler

import (
	"runtime"
	"time"

	"github.com/getlantern/systray"
	"github.com/kloudlite/kl/app/handler/icons"
	ns "github.com/kloudlite/kl/app/handler/name-conts"
	"github.com/kloudlite/kl/constants"
	"github.com/kloudlite/kl/domain/client"
	fn "github.com/kloudlite/kl/pkg/functions"
)

type projectMap map[string]map[string]string

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

	RedrawEnvs(projectMap)

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

type env struct {
	name   string
	envBtn *systray.MenuItem
}

type project struct {
	name       string
	projectBtn *systray.MenuItem
	envs       []env
}

type handler struct {
	channel              chan ChanelMsg
	itemMap              map[ns.ItemName]*systray.MenuItem
	projects             []project
	projectRenderVersion string
}

func NewHandler(channel chan ChanelMsg) Handler {
	return &handler{
		channel:  channel,
		itemMap:  make(map[ns.ItemName]*systray.MenuItem),
		projects: []project{},
	}
}

func (h *handler) Channel() chan ChanelMsg {
	return h.channel
}

func (h *handler) ItemMap() map[ns.ItemName]*systray.MenuItem {
	return h.itemMap
}

func (h *handler) ReconMeta() {
	systray.SetIcon(icons.Loading)
	// systray.SetTitle("Kloudlite")
	systray.SetTooltip("Kloudlite vpn client")

	go func() {
		for {
			b, err := client.IsLoading()
			if err != nil {
				fn.PrintError(err)
				fn.Alert("Error", err.Error())
			}

			if b {
				systray.SetTemplateIcon(icons.Loading, icons.Loading)
			} else {
				data, err := client.GetExtraData()
				if err != nil {
					data.VpnConnected = false
				}

				switch runtime.GOOS {
				case constants.RuntimeDarwin:
					if data.VpnConnected {
						systray.SetTemplateIcon(icons.Logo, icons.Logo)
					} else {
						systray.SetIcon(icons.DisabledLogo)
					}
				case constants.RuntimeLinux:
					if data.VpnConnected {
						systray.SetIcon(icons.Logo)
					} else {
						systray.SetIcon(icons.DisabledLogo)
					}
				}

			}

			<-time.After(time.Millisecond * 500)
		}
	}()
}

func (h *handler) ReconQuit() {
	mQuitOrig := systray.AddMenuItem("Quit", "Quit client")
	go func() {
		<-mQuitOrig.ClickedCh
		h.channel <- ChanelMsg{
			Msg:    "Device clicked",
			Item:   mQuitOrig,
			Action: ns.Quit,
		}
	}()
}
