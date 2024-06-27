package client

import (
	"os"
	"path"

	"github.com/kloudlite/kl/pkg/functions"
)

func Logout() error {
	configFolder, err := GetConfigFolder()
	if err != nil {
		return functions.NewE(err)
	}
	sessionFile, err := os.Stat(path.Join(configFolder, SessionFileName))
	if err != nil && os.IsNotExist(err) {
		return functions.Error("not logged in")
	}
	if err != nil {
		return functions.NewE(err)
	}

	accountFile, _ := os.Stat(path.Join(configFolder, MainCtxFileName))
	if accountFile != nil {
		if err := os.Remove(path.Join(configFolder, accountFile.Name())); err != nil {
			return functions.NewE(err)
		}
	}

	extraDataFile, _ := os.Stat(path.Join(configFolder, ExtraDataFileName))
	if extraDataFile != nil {
		if err := os.Remove(path.Join(configFolder, extraDataFile.Name())); err != nil {
			return functions.NewE(err)
		}
	}

	wgpidFile, _ := os.Stat(path.Join(configFolder, "wgpid"))
	if wgpidFile != nil {
		if err := os.Remove(path.Join(configFolder, wgpidFile.Name())); err != nil {
			return functions.NewE(err)
		}
	}

	// deviceFile, _ := os.Stat(path.Join(configFolder, DeviceFileName))
	// if deviceFile != nil {
	// 	if err := os.Remove(path.Join(configFolder, deviceFile.Name())); err != nil {
	// 		return functions.NewE(err)
	// 	}
	// }

	return os.Remove(path.Join(configFolder, sessionFile.Name()))
	//return os.RemoveAll(configFolder)
}
