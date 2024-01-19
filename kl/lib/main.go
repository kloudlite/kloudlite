package lib

import (
	"github.com/kloudlite/kl/domain/client"
)

func SelectProject(projectId string) error {

	file, err := client.GetAccountContext()
	if err != nil {
		return err
	}

	file.AccountName = projectId

	err = client.WriteAccountContext(file.AccountName)
	return err

}

func SelectDevice(deviceId string) error {

	file, err := client.GetAccountContext()
	if err != nil {
		return err
	}

	file.AccountName = deviceId

	err = client.WriteAccountContext(file.AccountName)
	return err

}
