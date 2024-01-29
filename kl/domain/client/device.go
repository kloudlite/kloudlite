package client

import (
	"errors"
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
