package handler

import (
	"fmt"

	ns "github.com/kloudlite/kl/app/handler/name-conts"
	"github.com/kloudlite/kl/constants"
	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/domain/server"
	fn "github.com/kloudlite/kl/pkg/functions"
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
						ac, err := client.GetAccountContext()
						if err != nil {
							fn.Println(err)
							fn.Alert("Open account settings failed", err.Error())
						}

						if err := open.Run(fmt.Sprintf("%s/%s/%s",
							constants.ConsoleUrl, ac.AccountName, "settings"),
						); err != nil {
							fn.Println(err)
							fn.Alert("Open account settings failed", err.Error())
						}

					}

				case ns.SwitchAccount:
					{
						if err := client.SelectAccount(a.Msg); err != nil {
							fn.Println(err)
							fn.Alert("Switch account failed", err.Error())
						}

						h.ReconAccount()
					}

				case ns.ToggleDevice:
					{
					}

					// action block end
				}
				// select block end
			}
		}
	}()
}
