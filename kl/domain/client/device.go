package client

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"time"

	proxy "github.com/kloudlite/kl/domain/dev-proxy"
	"github.com/kloudlite/kl/flags"
)

func CurrentDeviceName() (string, error) {
	file, err := GetDeviceContext()
	if err != nil {
		return "", err
	}
	if file.DeviceName == "" {
		return "",
			errors.New("no selected device. please select one using \"kl account switch\"")
	}
	return file.DeviceName, nil
}

func CurrentDeviceDNS() (*string, error) {
	dev, err := CurrentDeviceName()
	if err != nil {
		return nil, err
	}

	ipAddr, err := net.ResolveIPAddr("", fmt.Sprintf("%s.device.local", dev))
	if err != nil {
		return nil, err
	}

	kk := ipAddr.IP.String()
	return &kk, nil
}

func SelectDevice(deviceName string) error {
	file, err := GetDeviceContext()
	if err != nil {
		return err
	}

	file.DeviceName = deviceName

	if file.DeviceName == "" {
		return nil
	}

	err = WriteDeviceContext(deviceName)
	return err
}

func EnsureAppRunning() error {
	p, err := proxy.NewProxy(flags.IsDev(), false)
	if err != nil {
		return err
	}

	count := 0
	for {
		if p.Status() {
			return nil
		}

		// configFolder, err := client.GetConfigFolder()
		// if err != nil {
		// 	return err
		// }
		//
		// b, err := os.ReadFile(configFolder + "/apppid")
		//
		// if err == nil {
		// 	pid := string(b)
		// 	if fn.ExecCmd(fmt.Sprintf("ps -p %s", pid), nil, false) == nil {
		// 		return nil
		// 	}
		// }

		cmd := exec.Command("sudo", "echo", "")
		cmd.Stdin = os.Stdin
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout

		err := cmd.Run()
		if err != nil {
			return err
		}

		command := exec.Command("sudo", flags.CliName, "start-app")

		_ = command.Start()

		// err = os.WriteFile(configFolder+"/apppid", []byte(fmt.Sprintf("%d", command.Process.Pid)), 0644)
		// if err != nil {
		// 	fn.PrintError(err)
		// 	return err
		// }

		count++
		if count >= 10 {
			return fmt.Errorf("failed to start app")
		}

		time.Sleep(2 * time.Second)
	}
}
