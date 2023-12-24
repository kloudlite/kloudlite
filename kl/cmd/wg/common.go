package wg

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/kloudlite/kl/lib/common"
	"github.com/kloudlite/kl/lib/server"
	"github.com/kloudlite/kl/lib/wgc"
	"golang.zx2c4.com/wireguard/wgctrl"
)

func getDeviceSelect() (*server.Device, error) {

	deviceId, err := server.CurrentDeviceId()
	if err != nil {
		return nil, err
	}

	devices, err := server.GetDevices()
	if err != nil {
		return nil, err
	}

	for _, d := range devices {
		if d.Id == deviceId {
			return &d, err
		}
	}
	return nil, errors.New("plese select a device first using \"kl use device\"")

}

func startConfiguration(verbose bool) error {
	devices, err := server.GetDevices()
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

	if device.Region == "" {
		return errors.New("region not selected in device please use 'kl use device' to select device")
	}

	err = configure(device.Id, verbose)
	return err
}

func configure(
	deviceId string,
	verbose bool,
) error {

	s := common.NewSpinner()
	cfg := wgc.Config{}

	device, err := server.GetDevice(deviceId)
	if err != nil {
		return err
	}

	// time.Sleep(time.Second * 2)

	configuration := device.Configuration["config-"+device.Region]

	s.Start()
	if verbose {
		common.Log("[#] validating configuration")
	}
	if e := cfg.UnmarshalText([]byte(configuration)); e != nil {
		return e
	}
	s.Stop()

	if len(cfg.Address) == 0 {
		return errors.New("device ip not found")
	} else if e := setDeviceIp(cfg.Address[0].IP.String(), verbose); e != nil {
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
		common.Log("[#] setting up connection")
	}

	err = wg.ConfigureDevice(KlWgInterface, cfg.Config)
	if err != nil {
		fmt.Printf("failed to configure device: %v", err)
	}

	for _, i2 := range cfg.Peers[0].AllowedIPs {
		err = ipRouteAdd(i2.String(), cfg.Address[0].IP.String(), verbose)
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
		common.Log("[#] " + strings.Join(cmdArr, " "))
		cmd.Stdout = os.Stdout
	}
	cmd.Stderr = os.Stderr
	// s.Start()
	err = cmd.Run()
	// s.Stop()
	return err
}
