package main

import (
	"fmt"
	"io/fs"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2"
)

type Service struct {
	Name     string `json:"name"`
	Target   int    `json:"servicePort"`
	Port     int    `json:"proxyPort"`
	Listener net.Listener
	Closed   bool
}

const (
	WgFileName          = "wg0"
	WgFileNameSecondary = "sample"
	WgFile              = "/etc/wireguard/" + WgFileName + ".conf"
	WgFileSecondary     = "/etc/wireguard/" + WgFileNameSecondary + ".conf"
)

func reloadConfig(conf []byte) error {
	isFirstTime := conf == nil
	if conf == nil {
		var err error
		conf, err = os.ReadFile(WgFile)
		if err != nil {
			return err
		}
	}

	err := os.WriteFile(WgFileSecondary, conf, fs.ModeAppend)
	if err != nil {
		return err
	}

	if isFirstTime {
		fmt.Println("[#] Wireguard Server Starting")

		cmds := strings.Fields("wg-quick up " + WgFileNameSecondary)

		cmd := exec.Command(cmds[0], cmds[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin
		cmd.Stderr = os.Stderr

		err = cmd.Run()
		if err != nil {
			return err
		}

		fmt.Println("[#] Wireguard Server Started")
		return err
	}

	fmt.Println("[#] Wireguard Server Restarting")
	// cmds := strings.Fields("wg-quick strip " + WgFileNameSecondary)
	cmd := exec.Command("bash", "-c", fmt.Sprintf("wg-quick strip %s > a.txt && wg syncconf %s a.txt", WgFileNameSecondary, WgFileNameSecondary))
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		fmt.Println(err)
	}

	if err != nil {
		return err
	}

	fmt.Println("[#] Wireguard Server Restarted")
	return nil
}

func startApi() error {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	app.Post("/post", func(c *fiber.Ctx) error {
		err := reloadConfig(c.Body())
		if err != nil {
			return err
		}
		err = c.Send([]byte("done"))
		if err != nil {
			return err
		}
		return nil
	})

	err := app.Listen(":2998")
	if err != nil {
		return err
	}
	return nil
}

func main() {
	go func() {
		_ = startApi()
	}()
	err := reloadConfig(nil)
	if err != nil {
		panic(err)
	}
	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
