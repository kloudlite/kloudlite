package client

import (
	"errors"
)

//func SelectDevice(devName string) error {
//	file, err := GetDeviceContext()
//	if err != nil {
//		return err
//	}
//
//	file.DeviceName = devName
//
//	err = WriteDeviceContext(devName)
//	return err
//}

func SelectInfraDevice(devName string) error {
	file, err := GetActiveInfraContext()
	if err != nil {
		return err
	}

	file.DeviceName = devName

	err = WriteInfraContextFile(*file)
	return err
}

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

func CurrentInfraDeviceName() (string, error) {
	file, err := GetActiveInfraContext()
	if err != nil {
		return "", err
	}
	if file.DeviceName == "" {
		return "",
			errors.New("no selected device. please select one using \"kl infra vpn switch\"")
	}
	return file.DeviceName, nil
}
