package wg_svc

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"

	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/flags"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/spinner"
	"github.com/kloudlite/kl/pkg/zip"
)

func ensureInstalled() error {
	configFolder, err := client.GetConfigFolder()
	if err != nil {
		return err
	}

	appPath := path.Join(configFolder, "app")

	if _, err := os.Stat(appPath); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(appPath, 0755); err != nil {
				return err
			}
		}
	}

	if _, err := os.Stat(path.Join(appPath, "kloudlite")); err != nil {
		if os.IsNotExist(err) {
			return installApp()
		}
	}

	return nil
}

func installApp() error {

	fn.Log(fmt.Sprintf("[#] downloading app version %s", flags.Version))

	success := false

	s := spinner.NewSpinner()
	s.Start()

	defer func() {
		s.Stop()
		if success {
			fn.Log(fmt.Sprintf("[#] app version %s downloaded successfully", flags.Version))
		} else {
			fn.Log(fmt.Sprintf("[#] failed to download app version %s", flags.Version))
		}
	}()

	specUrl := fmt.Sprint("https://github.com/kloudlite/vpn-app/release/", flags.Version, ".zip")
	resp, err := http.Get(specUrl)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {

		return fmt.Errorf("failed to download app, status code: %d", resp.StatusCode)
	}

	configFolder, err := client.GetConfigFolder()
	if err != nil {
		return err
	}

	out, err := os.Create(path.Join(configFolder, "app", "kloudlite.zip"))
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, resp.Body); err != nil {
		return err
	}

	if err := zip.Unzip(path.Join(configFolder, "app", "kloudlite.zip")); err != nil {
		return err
	}

	_ = os.Remove(path.Join(configFolder, "app", "kloudlite.zip"))

	success = true
	return nil
}

func startApp() error {
	configFolder, err := client.GetConfigFolder()
	if err != nil {
		return err
	}

	appPath := path.Join(configFolder, "app", "kloudlite.app")

	if _, err := os.Stat(appPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("app not installed")
		}
	}

	if err := os.Chmod(appPath, 0755); err != nil {
		return err
	}

	if err := fn.ExecCmd(fmt.Sprintf("start %s", appPath), nil, false); err != nil {
		return err
	}

	return nil
}
