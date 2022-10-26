package main

import (
	"fmt"
	"io/fs"
	"io/ioutil"
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
	WgFile          = "/etc/wireguard/" + WgFileName + ".conf"
	WgFileSecondary = "/etc/wireguard/" + WgFileNameSecondary + ".conf"
)

func reloadConfig(conf []byte) error {
	fmt.Println("\n================== Restart ==================")
	if conf == nil {
		var err error
		conf, err = os.ReadFile(WgFile)
		if err != nil {
			return err
		}
	}
	// cmds := strings.Fields("chmod +rwx /etc/wireguard")
	// cmd := exec.Command(cmds[0], cmds[1:]...)
	// cmd.Run()

	err := ioutil.WriteFile(WgFileSecondary, conf, fs.ModeAppend)
	if err != nil {
		return err
	}

	cmds := strings.Fields("wg-quick down " + WgFileNameSecondary)

	cmd := exec.Command(cmds[0], cmds[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		fmt.Println(err)
	}

	cmds = strings.Fields("wg-quick up " + WgFileNameSecondary)

	cmd = exec.Command(cmds[0], cmds[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	err = cmd.Run()

	return err
}

func startApi() error {
	app := fiber.New()
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
