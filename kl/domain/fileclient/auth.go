package fileclient

import (
	fn "github.com/kloudlite/kl/pkg/functions"
	"os"
	"path"
)

func (fc *fclient) Logout() error {
	configPath := fc.configPath

	sessionFile, err := os.Stat(path.Join(configPath, SessionFileName))
	if err != nil && os.IsNotExist(err) {
		return fn.Error("not logged in")
	}
	if err != nil {
		return fn.NewE(err)
	}

	extraDataFile, _ := os.Stat(path.Join(configPath, ExtraDataFileName))
	if extraDataFile != nil {
		if err := os.Remove(path.Join(configPath, extraDataFile.Name())); err != nil {
			return fn.NewE(err)
		}
	}
	hashConfigPath := path.Join(configPath, "box-hash")

	_, err = os.Stat(hashConfigPath)
	if err == nil {
		if err = os.RemoveAll(hashConfigPath); err != nil {
			return fn.NewE(err)
		}
	}

	k3sConfigPath := path.Join(configPath, "k3s-local")

	_, err = os.Stat(k3sConfigPath)
	if err == nil {
		if err := os.RemoveAll(k3sConfigPath); err != nil {
			return fn.NewE(err)
		}
	}

	vpnConfigPath := path.Join(configPath, "vpn")
	_, err = os.Stat(vpnConfigPath)
	if err == nil {
		if err := os.RemoveAll(vpnConfigPath); err != nil {
			return fn.NewE(err)
		}
	}

	return os.Remove(path.Join(configPath, sessionFile.Name()))
}
