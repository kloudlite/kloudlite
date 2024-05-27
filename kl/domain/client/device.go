package client

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"syscall"
	"time"
	"unsafe"

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

func CurrentDeviceIp() (*string, error) {
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

	err = WriteDeviceContext(&DeviceContext{
		DeviceName: deviceName,
	})
	return err
}

func EnsureAppRunning() error {
	p, err := proxy.NewProxy(flags.IsDev())
	if err != nil {
		return err
	}

	count := 0
	for {
		if p.Status() {
			return nil
		}

		if runtime.GOOS != "windows" {
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

		} else {
			command := exec.Command(flags.CliName, "start-app")

			_ = command.Start()
		}

		count++
		if count >= 10 {
			return fmt.Errorf("failed to start app")
		}

		time.Sleep(2 * time.Second)
	}
}

var (
	shell32           = syscall.NewLazyDLL("shell32.dll")
	procShellExecuteW = shell32.NewProc("ShellExecuteW")
)

func ShellExecute(operation, file, parameters, directory string, showCmd int) error {
	op, err := syscall.UTF16PtrFromString(operation)
	if err != nil {
		return err
	}
	f, err := syscall.UTF16PtrFromString(file)
	if err != nil {
		return err
	}
	p, err := syscall.UTF16PtrFromString(parameters)
	if err != nil {
		return err
	}
	d, err := syscall.UTF16PtrFromString(directory)
	if err != nil {
		return err
	}
	ret, _, _ := procShellExecuteW.Call(
		0,
		uintptr(unsafe.Pointer(op)),
		uintptr(unsafe.Pointer(f)),
		uintptr(unsafe.Pointer(p)),
		uintptr(unsafe.Pointer(d)),
		uintptr(showCmd),
	)
	if ret <= 32 {
		return syscall.GetLastError()
	}
	return nil
}
