/*
Copyright © 2022 Kloudlite <support@kloudlite.io>

*/
package main

import (
	"fmt"
	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/ipc"
	"golang.zx2c4.com/wireguard/tun"
	"kloudlite.io/cmd/internal/cmd"
	"os"
)

func main() {
	cmd.Execute()
}

func main2() {
	var err error
	tun, err := func() (tun.Device, error) {
		return tun.CreateTUN("utun1729", device.DefaultMTU)
	}()
	if err != nil {
		fmt.Println(err)
		return
	}

	fileUAPI, err := func() (*os.File, error) {
		return ipc.UAPIOpen("utun1729")
	}()

	if err != nil {
		fmt.Println(err)
		return
	}

	uapi, err := ipc.UAPIListen("utun1729", fileUAPI)
	device := device.NewDevice(tun, conn.NewDefaultBind(), nil)
	errs := make(chan error)
	go func() {
		for {
			conn, e := uapi.Accept()
			if e != nil {
				errs <- e
				return
			}
			go device.IpcHandle(conn)
		}
	}()
	fmt.Println(err)

}
