package fileclient

import (
	"encoding/json"
	"fmt"
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
	if err = os.RemoveAll(hashConfigPath); err != nil {
		return fn.NewE(err)
	}
	vpnConfigPath := path.Join(configPath, "vpn")
	files, err := os.ReadDir(vpnConfigPath)
	if err != nil {
		return fn.NewE(err)
	}
	for _, file := range files {
		_, err := os.Stat(path.Join(vpnConfigPath, file.Name()))
		if err != nil {
			fn.PrintError(err)
			continue
		}
		content, err := os.ReadFile(path.Join(vpnConfigPath, file.Name()))
		if err != nil {
			fn.PrintError(err)
			continue
		}

		var data AccountVpnConfig
		err = json.Unmarshal(content, &data)
		if err != nil {
			fn.PrintError(err)
			continue
		}
		data.WGconf = ""

		modifiedContent, err := json.Marshal(data)
		if err != nil {
			fmt.Println(err)
			continue
		}

		err = os.WriteFile(path.Join(vpnConfigPath, file.Name()), modifiedContent, 0644)
		if err != nil {
			fmt.Println(err)
		}
	}

	return os.Remove(path.Join(configPath, sessionFile.Name()))
}
