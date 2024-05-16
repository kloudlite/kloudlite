package app

import (
	"github.com/getlantern/systray"
	"github.com/kloudlite/kl/app/handler"
	"github.com/kloudlite/kl/app/server"
	fn "github.com/kloudlite/kl/pkg/functions"
)

func RunApp(binName string) error {
	onExit := func() {
		fn.Log("Exiting...")
		// now := time.Now()
		// ioutil.WriteFile(fmt.Sprintf(`on_exit_%d.txt`, now.UnixNano()), []byte(now.String()), 0644)
	}

	go func() {
		s := server.New(binName)
		s.Start()
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
	h.ReconAccount()
	systray.AddSeparator()

	// handle actions releated to device
	h.ReconDevice()
	systray.AddSeparator()

	// handle actions releated to environment
	// h.ReconEnv()
	// systray.AddSeparator()

	// h.ReconUser()
	// h.ReconAuth()

	h.ReconQuit()
	h.StartListener()
}
