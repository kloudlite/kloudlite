package handler

import (
	"fmt"

	"github.com/getlantern/systray"
	ns "github.com/kloudlite/kl/app/handler/name-conts"
	"github.com/kloudlite/kl/constants"
	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/domain/server"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/wg_vpn/wgc"
	"github.com/skratchdot/open-golang/open"
)

func (h *handler) StartListener() {
	go func() {
		for {
			select {
			case a := <-h.channel:
				switch a.Action {
				case ns.Login:
					{
						loginId, err := server.CreateRemoteLogin()
						if err != nil {
							fn.Println(err)
							fn.Alert("Login failed", err.Error())
							continue
						}

						if err := open.Run(fmt.Sprintf("%s/%s%s", constants.LoginUrl, "?loginId=", loginId)); err != nil {
							fn.Println(err)
							fn.Alert("Login failed", err.Error())
							continue
						}

						if err := server.Login(loginId); err != nil {
							fn.Println(err)
							fn.Alert("Login failed", err.Error())
						}

						h.ReconAuth()
						h.ReconUser()
					}

				case ns.Logout:
					{
						fmt.Println("Sign out clicked")
						if err := client.Logout(); err != nil {
							fn.Println(err)
							fn.Alert("Sign out failed", err.Error())
						}
						h.ReconAuth()
						h.ReconUser()
					}

				case ns.OpenAccountSettings:
					{
						//ac, err := client.GetMainCtx()
						//if err != nil {
						//	fn.Println(err)
						//	fn.Alert("Open account settings failed", err.Error())
						//}

						//if err := open.Run(fmt.Sprintf("%s/%s/%s",
						//	constants.ConsoleUrl, ac.AccountName, "settings"),
						//); err != nil {
						//	fn.Println(err)
						//	fn.Alert("Open account settings failed", err.Error())
						//}

					}

				case ns.SwitchAccount:
					{
						if err := client.SelectAccount(a.Msg); err != nil {
							fn.Println(err)
							fn.Alert("Switch account failed", err.Error())
						}

						h.ReconAccount()
						h.ReconEnv()
					}

				case ns.ToggleDevice:
					{
						func() {

							err := client.SetLoading(true)
							if err == nil {
								defer client.SetLoading(false)
							}

							cmd := constants.InfraCliName
							if s, err := client.CurrentClusterName(); err != nil || s == "" {
								kt, err := client.GetKlFile("")
								defEnv := ""
								if err == nil && kt.DefaultEnv != "" {
									defEnv = kt.DefaultEnv
								}

								if e, err := client.CurrentEnv(); (err != nil || e == nil || e.Name == "") && (defEnv == "") {
									fn.Println(err)
									fn.Alert("No Cluster or Environment Selected", "Please select a cluster or environment")
									return
								}

								cmd = constants.CoreCliName
							}

							if !IsCmdExists(cmd, true) {
								return
							}

							if !server.CheckDeviceStatus() {
								if err := fn.ExecCmd(fmt.Sprintf("%s vpn start -s", cmd), nil, true); err != nil {
									fn.PrintError(err)
									fn.Alert("Start VPN failed", err.Error())
								}
								fn.Notify("Info", "Kloudlite VPN Connected")
								return
							}

							if err := fn.ExecCmd(fmt.Sprintf("%s vpn stop", cmd), nil, true); err != nil {
								fn.PrintError(err)
								fn.Alert("Stop VPN failed", err.Error())
							}
							fn.Notify("Info", "Kloudlite VPN Disconnected")
						}()
					}

				case ns.Quit:
					{

						wgInterface, err := wgc.Show(&wgc.WgShowOptions{
							Interface: "interfaces",
						})

						if err != nil {
							fn.PrintError(err)
							fn.Notify("error:", err)
						}

						if len(wgInterface) > 0 {

							cmd := "kl"

							if !IsCmdExists(cmd, false) {
								cmd = "kli"
								if !IsCmdExists(cmd, true) {
									systray.Quit()
									continue
								}
							}

							if err := fn.ExecCmd(fmt.Sprintf("%s vpn stop", cmd), nil, true); err != nil {
								fn.PrintError(err)
								fn.Alert("Stop VPN failed", err.Error())
							}
						}

						systray.Quit()
					}

					// action block end
				}
				// select block end
			}
		}
	}()
}

func IsCmdExists(cmd string, verbose bool) bool {
	if fn.ExecCmd(fmt.Sprintf("whereis %s", cmd), nil, false) != nil {
		if verbose {
			fn.Println(fmt.Sprintf("%s not found", cmd))
			fn.Alert("Cmd not found", fmt.Sprintf("%s not found", cmd))
		}
		return false
	}

	return true
}
