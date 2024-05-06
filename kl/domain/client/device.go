package client

import (
	"errors"
	"fmt"
	"net"
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

func CurrentDeviceDNS() ([]net.IP, error) {
	dev, err := CurrentDeviceName()
	if err != nil {
		return nil, err
	}
	ips, err := net.LookupIP(fmt.Sprintf("%s.local", dev))
	if err != nil {
		return nil, err
	}
	return ips, nil
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
