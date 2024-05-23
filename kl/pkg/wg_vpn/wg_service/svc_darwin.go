package wg_svc

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"time"

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

	if _, err := os.Stat(path.Join(appPath, "kloudlite.app")); err != nil {
		if os.IsNotExist(err) {
			return installApp()
		}
	}

	return nil
}

func installApp() error {

	fn.Log(fmt.Sprintf("[#] downloading app version %s", flags.Version))

	success := false

	defer spinner.Client.Start()()
	defer func() {
		if success {
			fn.Log(fmt.Sprintf("[#] app version %s downloaded successfully", flags.Version))
		} else {
			fn.Log(fmt.Sprintf("[#] failed to download app version %s", flags.Version))
		}
	}()

	specUrl := fmt.Sprint("https://github.com/kloudlite/vpn-app/releases/download/", flags.Version, "/kloudlite_macos.zip")
	fmt.Println(specUrl)
	resp, err := http.Get(specUrl)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		// using fallback url

		specUrl = fmt.Sprint("https://github.com/kloudlite/vpn-app/releases/download/v1.0.5-nightly/kloudlite_macos.zip")
		var err2 error
		resp, err2 = http.Get(specUrl)
		if err2 != nil {
			return err2
		}
	}

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

	if err := zip.Unzip(path.Join(configFolder, "app", "kloudlite.zip"), path.Join(configFolder, "app")); err != nil {
		return err
	}

	_ = os.Remove(path.Join(configFolder, "app", "kloudlite.zip"))

	success = true
	return nil
}

func startApp() error {

	fn.Log("[#] starting service")
	success := false
	defer func() {
		if success {
			fn.Log("[#] service started successfully")
		} else {
			fn.Log("[#] failed to start service")
		}
	}()

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

	if err := fn.ExecCmd(fmt.Sprintf("open %s", appPath), nil, false); err != nil {
		return err
	}

	count := 0
	for {
		if count == 5 || isReady() {
			break
		}
		time.Sleep(1 * time.Second)
		count += 1
	}

	if isReady() {
		success = true
		return nil
	}

	return fmt.Errorf("failed to start service")
}
