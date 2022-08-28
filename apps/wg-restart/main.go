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
	WG_FILE_NAME           = "wg0"
	WG_FILE_NAME_SECONDARY = "sample"
	WG_FILE                = "/etc/wireguard/" + WG_FILE_NAME + ".conf"
	WG_FILE_SECONDARY      = "/etc/wireguard/" + WG_FILE_NAME_SECONDARY + ".conf"
)

func reloadConfig(conf []byte) error {
	fmt.Println("\n================== Restart ==================")
	if conf == nil {
		var err error
		conf, err = ioutil.ReadFile(WG_FILE)
		if err != nil {
			return err
		}
	}
	// cmds := strings.Fields("chmod +rwx /etc/wireguard")
	// cmd := exec.Command(cmds[0], cmds[1:]...)
	// cmd.Run()

	err := ioutil.WriteFile(WG_FILE_SECONDARY, conf, fs.ModeAppend)
	if err != nil {
		return err
	}

	cmds := strings.Fields("wg-quick down " + WG_FILE_NAME_SECONDARY)

	cmd := exec.Command(cmds[0], cmds[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		fmt.Println(err)
	}

	cmds = strings.Fields("wg-quick up " + WG_FILE_NAME_SECONDARY)

	cmd = exec.Command(cmds[0], cmds[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	err = cmd.Run()

	return err
}

func startApi() {
	app := fiber.New()
	app.Post("/post", func(c *fiber.Ctx) error {
		err := reloadConfig(c.Body())
		if err != nil {
			return err
		}
		c.Send([]byte("done"))
		return nil
	})
	app.Listen(":2998")
}
func main() {
	go startApi()
	err := reloadConfig(nil)
	if err != nil {
		panic(err)
	}
	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
