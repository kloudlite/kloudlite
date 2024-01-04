package wg

import (
	"encoding/base64"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/domain/server"
	"github.com/kloudlite/kl/lib/wgc"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/spinner"
	"golang.zx2c4.com/wireguard/wgctrl"
)

func getDeviceSelect() (*server.Device, error) {

	devName, err := client.CurrentDeviceName()
	if err != nil {
		return nil, err
	}

	devices, err := server.ListDevices()
	if err != nil {
		return nil, err
	}

	for _, d := range devices {
		if d.Metadata.Name == devName {
			return &d, err
		}
	}
	return nil, errors.New("plese select a device first using \"kl use device\"")

}

func startConfiguration(verbose bool) error {
	devices, err := server.ListDevices()
	if err != nil {
		return err
	}
	if len(devices) == 0 {
		return errors.New("no Devices found")
	}
	device, err := getDeviceSelect()
	if err != nil {
		return err
	}

	if runtime.GOOS == "darwin" {
		return configureDarwin(device.Metadata.Name, verbose)
	}

	return configure(device.Metadata.Name, device.Metadata.Name, verbose)
}

func configure(
	devName string,
	interfaceName string,
	verbose bool,
) error {

	s := spinner.NewSpinner()
	cfg := wgc.Config{}

	device, err := server.GetDevice(fn.MakeOption("deviceName", devName))
	if err != nil {
		return err
	}

	// time.Sleep(time.Second * 2)
	if device.WireguardConfig == nil {
		return errors.New("no wireguard config found")
	}

	configuration, err := base64.StdEncoding.DecodeString(device.WireguardConfig.Value)
	if err != nil {
		return err
	}

	s.Start()
	if verbose {
		fn.Log("[#] validating configuration")
	}
	if e := cfg.UnmarshalText([]byte(configuration)); e != nil {
		return e
	}
	s.Stop()

	if len(cfg.Address) == 0 {
		return errors.New("device ip not found")
	} else if e := setDeviceIp(cfg.Address[0].IP.String(), devName, verbose); e != nil {
		return e
	}

	if e := setDNS(cfg.DNS, verbose); e != nil {
		return e
	}

	wg, err := wgctrl.New()
	if err != nil {
		return err
	}

	if verbose {
		fn.Log("[#] setting up connection")
	}

	err = wg.ConfigureDevice(interfaceName, cfg.Config)
	if err != nil {
		fmt.Printf("failed to configure device: %v", err)
	}

	for _, i2 := range cfg.Peers[0].AllowedIPs {
		err = ipRouteAdd(i2.String(), cfg.Address[0].IP.String(), interfaceName, verbose)
		if err != nil {
			return err
		}
	}

	return err
}

func execCmd(cmdString string, verbose bool) error {
	r := csv.NewReader(strings.NewReader(cmdString))
	r.Comma = ' '
	cmdArr, err := r.Read()
	if err != nil {
		return err
	}
	cmd := exec.Command(cmdArr[0], cmdArr[1:]...)
	if verbose {
		fn.Log("[#] " + strings.Join(cmdArr, " "))
		cmd.Stdout = os.Stdout
	}
	cmd.Stderr = os.Stderr
	// s.Start()
	err = cmd.Run()
	// s.Stop()
	return err
}
