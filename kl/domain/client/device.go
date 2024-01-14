package client

import "errors"

func SelectDevice(devName string) error {
	file, err := GetActiveContext()
	if err != nil {
		return err
	}

	file.DeviceName = devName

	err = WriteContextFile(*file)
	return err
}

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
	file, err := GetActiveContext()
	if err != nil {
		return "", err
	}
	if file.DeviceName == "" {
		return "",
			errors.New("no selected device. please select one using \"kl infra vpn switch\"")
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
