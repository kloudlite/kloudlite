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
