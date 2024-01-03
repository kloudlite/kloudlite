package lib

import (
	"github.com/kloudlite/kl/domain/client"
)

func SelectProject(projectId string) error {

	file, err := client.GetContextFile()
	if err != nil {
		return err
	}

	file.ProjectName = projectId

	err = client.WriteContextFile(*file)
	return err

}

func SelectDevice(deviceId string) error {

	file, err := client.GetContextFile()
	if err != nil {
		return err
	}

	file.DeviceName = deviceId

	err = client.WriteContextFile(*file)
	return err

}
