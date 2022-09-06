package wg

import (
	"errors"
	"fmt"
	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/ipc"
	"golang.zx2c4.com/wireguard/tun"
	"kloudlite.io/cmd/internal/lib/common"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
)

func configure(
	dnsIp string,
	deviceIp string,
	endpoint string,
	privateKey string,
	publicKey string,
) error {
	configFolder, err := common.GetConfigFolder()
	if _, err := os.Stat(fmt.Sprintf(configFolder + "/wgpid")); errors.Is(err, os.ErrNotExist) {
		startService()
		return configure(dnsIp, deviceIp, endpoint, privateKey, publicKey)
	}
	err = setDNS(dnsIp)
	if err != nil {
		return err
	}
	err = setDeviceIp(deviceIp)
	if err != nil {
		return err
	}
	return nil
}

func setDNS(dnsIp string) error {
	file, err := os.ReadFile("/etc/resolv.conf")
	if err != nil {
		return err
	}
	if _, err := os.Stat("/etc/resolv.conf.back"); errors.Is(err, os.ErrNotExist) {
		err = os.WriteFile("/etc/resolv.conf.back", file, 0644)
		if err != nil {
			return err
		}
	}
	err = os.WriteFile("/etc/resolv.conf", []byte("nameserver "+dnsIp), 0644)
	if err != nil {
		return err
	}
	return nil
}

func resetDNS(dnsIp string) error {
	err := os.Remove("/etc/resolv.conf")
	if err != nil {
		return err
	}
	err = os.Rename("/etc/resolv.conf.back", "/etc/resolv.conf")
	if err != nil {
		return err
	}
	return nil
}

func setDeviceIp(deviceIp string) error {
	return exec.Command("ifconfig", "utun1729", deviceIp, deviceIp).Run()
}

func startService() {
	t, err := tun.CreateTUN("utun1729", device.DefaultMTU)
	if err != nil {
		fmt.Println(err)
		return
	}
	fileUAPI, err := ipc.UAPIOpen("utun1729")
	if err != nil {
		fmt.Println(err)
		return
	}
	logger := device.NewLogger(
		device.LogLevelVerbose,
		"(utun1729) ",
	)
	d := device.NewDevice(t, conn.NewDefaultBind(), logger)
	logger.Verbosef("Device started")
	errs := make(chan error)
	term := make(chan os.Signal, 1)
	uapi, err := ipc.UAPIListen("utun1729", fileUAPI)
	if err != nil {
		logger.Errorf("Failed to listen on uapi socket: %v", err)
		os.Exit(1)
	}
	go func() {
		for {
			conn, err := uapi.Accept()
			if err != nil {
				errs <- err
				return
			}
			go d.IpcHandle(conn)
		}
	}()
	logger.Verbosef("UAPI listener started")
	signal.Notify(term, syscall.SIGTERM)
	signal.Notify(term, os.Interrupt)

	select {
	case <-term:
	case <-errs:
	case <-d.Wait():
	}
	uapi.Close()
	d.Close()
	logger.Verbosef("Shutting down")
}

func stopService() {
	configFolder, err := common.GetConfigFolder()
	if err != nil {
		return
	}
	file, err := os.ReadFile(configFolder + "/wgpid")
	if err != nil {
		return
	}
	i, err := strconv.ParseInt(string(file), 10, 64)
	if err != nil {
		return
	}
	err = syscall.Kill(int(i), syscall.SIGTERM)
	if err != nil {
		return
	}
	os.Remove(configFolder + "/wgpid")
}
