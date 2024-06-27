package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "logout from kloudlite",
	Example: `# Logout from kloudlite
{cmd} auth logout`,
	Run: func(*cobra.Command, []string) {
		configFolder, err := client.GetConfigFolder()
		if err != nil {
			fn.Log(err)
			return
		}

		if err = logout(configFolder); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func logout(configPath string) error {
	sessionFile, err := os.Stat(path.Join(configPath, client.SessionFileName))
	if err != nil && os.IsNotExist(err) {
		return fn.NewE(errors.New("not logged in"))
	}
	if err != nil {
		return functions.NewE(err)
	}

	extraDataFile, _ := os.Stat(path.Join(configPath, client.ExtraDataFileName))
	if extraDataFile != nil {
		if err := os.Remove(path.Join(configPath, extraDataFile.Name())); err != nil {
			return functions.NewE(err)
		}
	}
	hashConfigPath := path.Join(configPath, "box-hash")
	if err = os.RemoveAll(hashConfigPath); err != nil {
		return functions.NewE(err)
	}
	vpnConfigPath := path.Join(configPath, "vpn")
	files, err := os.ReadDir(vpnConfigPath)
	if err != nil {
		return functions.NewE(err)
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

		var data client.AccountVpnConfig
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
