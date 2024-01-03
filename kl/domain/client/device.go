package client

import "errors"

func SelectDevice(devName string) error {
	file, err := GetContextFile()
	if err != nil {
		return err
	}

	file.DeviceName = devName

	err = WriteContextFile(*file)
	return err
}

func CurrentDeviceName() (string, error) {
	file, err := GetContextFile()
	if err != nil {
		return "", err
	}
	if file.DeviceName == "" {
		return "", errors.New("noSelectedDevice")
	}
	if file.DeviceName == "" {
		return "",
			errors.New("no selected device. please select a device using \"kl use device\"")
	}
	return file.DeviceName, nil
}
