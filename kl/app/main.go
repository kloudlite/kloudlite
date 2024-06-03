package app

import (
	"context"
	"os"
	"path"
	"time"

	// "github.com/kloudlite/kl/app/handler"
	"github.com/kloudlite/kl/app/server"
	"github.com/kloudlite/kl/cmd/box/boxpkg"
	"github.com/kloudlite/kl/domain/client"
	fn "github.com/kloudlite/kl/pkg/functions"
)

func RunApp(binName string) error {
	fn.Log("kl vpn and proxy controller")

	// onExit := func() {
	// 	fn.Log("Exiting...")
	// }

	ctx, cf := context.WithCancel(context.Background())

	ch := make(chan error, 0)

	go func() {
		if err := func() error {
			configFolder, err := client.GetConfigFolder()
			if err != nil {
				return err
			}

			ipPath := path.Join(configFolder, "host_ip")
			for {
				if ctx.Err() != nil {
					return err
				}

				if err := func() error {
					s, err := boxpkg.GetDockerHostIp()
					if err != nil {
						return err
					}

					b, err := os.ReadFile(ipPath)
					if err == nil {
						if string(b) == s {
							return nil
						}
					}

					return os.WriteFile(ipPath, []byte(s), os.ModePerm)
				}(); err != nil {
					fn.Alert("Error", err.Error())
					time.Sleep(1 * time.Second)
					continue
				}

				time.Sleep(10 * time.Second)
			}
		}(); err != nil {
			ch <- err
			return
		}
	}()

	go func() {
		s := server.New(binName)
		ch <- s.Start(ctx)
	}()

	// systray.Run(func() {
	// 	onReady(binName)
	// }, onExit)

	select {
	case i := <-ch:
		cf()

		return i
	}

}

// func onReady(binName string) {
// 	channel := make(chan handler.ChanelMsg)
// 	h := handler.NewHandler(channel, binName)
//
// 	// setup logo and tooltip
// 	h.ReconMeta()
//
// 	// systray.AddSeparator()
//
// 	// handle actions releated to account
// 	// h.ReconAccount()
// 	// systray.AddSeparator()
//
// 	// handle actions releated to device
// 	// h.ReconDevice()
// 	// systray.AddSeparator()
//
// 	// handle actions releated to environment
// 	// h.ReconEnv()
// 	// systray.AddSeparator()
//
// 	// h.ReconUser()
// 	// h.ReconAuth()
//
// 	h.ReconQuit()
// 	h.StartListener()
// }
