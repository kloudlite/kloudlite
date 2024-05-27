//go:build windows

package app

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/kloudlite/kl/app/server"
	"github.com/kloudlite/kl/cmd/box/boxpkg"
	"github.com/kloudlite/kl/domain/client"
	fn "github.com/kloudlite/kl/pkg/functions"
)

func RunApp(binName string) error {

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

	s := server.New(binName)
	if err := s.Start(); err != nil {
		fn.PrintError(err)
	}

	return nil
}
