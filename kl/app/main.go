package app

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/getlantern/systray"
	"github.com/kloudlite/kl/app/handler"
	"github.com/kloudlite/kl/app/server"
	"github.com/kloudlite/kl/cmd/box/boxpkg"
	"github.com/kloudlite/kl/domain/client"
	fn "github.com/kloudlite/kl/pkg/functions"
)

func RunApp(binName string) error {
	onExit := func() {
		fn.Log("Exiting...")
		// now := time.Now()
		// ioutil.WriteFile(fmt.Sprintf(`on_exit_%d.txt`, now.UnixNano()), []byte(now.String()), 0644)
	}

	go func() {
		configFolder, err := client.GetConfigFolder()
		if err != nil {
			fmt.Println(err)
			fn.Alert("Error", err.Error())
			return
		}
		ipPath := path.Join(configFolder, "host_ip")
		for {
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
	}()

	go func() {
		s := server.New(binName)
		if err := s.Start(); err != nil {
			fn.PrintError(err)
		}
	}()

	systray.Run(func() {
		onReady(binName)
	}, onExit)

	return nil
}

func onReady(binName string) {
	channel := make(chan handler.ChanelMsg)
	h := handler.NewHandler(channel, binName)

	// setup logo and tooltip
	h.ReconMeta()

	// systray.AddSeparator()

	// handle actions releated to account
	// h.ReconAccount()
	// systray.AddSeparator()

	// handle actions releated to device
	// h.ReconDevice()
	// systray.AddSeparator()

	// handle actions releated to environment
	// h.ReconEnv()
	// systray.AddSeparator()

	// h.ReconUser()
	// h.ReconAuth()

	h.ReconQuit()
	h.StartListener()
}
