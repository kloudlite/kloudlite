package updater

import (
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/kloudlite/kl/flags"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
)

type updater struct {
	releaseInfo map[string]string
}

type Updater interface {
	CheckForUpdates() (bool, error)
	GetUpdateMessage() (*string, error)
	Update() error
	GetUpdateUrl() (*string, error)
}

func NewUpdater() Updater {
	return &updater{}
}

func (u *updater) GetUpdateUrl() (*string, error) {
	relInfo, err := u.fetchReleaseInfo()
	if err != nil {
		return nil, err
	}

	updateUrl, ok := relInfo["update_url"]
	if !ok {
		return nil, fn.Errorf("update url is not available")
	}

	b, err := base64.StdEncoding.DecodeString(updateUrl)
	if err != nil {
		return nil, fn.NewE(err, "failed to decode update url")
	}

	updateUrl = string(b)

	return &updateUrl, nil
}

func (u *updater) Update() error {
	relInfo, err := u.fetchReleaseInfo()
	if err != nil {
		fn.PrintError(err)
		return err
	}

	updateUrl, ok := relInfo["update_url"]
	if !ok {
		return fn.Errorf("update url is not available")
	}

	r, err := http.Get(updateUrl)
	if err != nil {
		return err
	}

	fn.Log(r)
	return nil
}

func parseTxtResp(data string) (map[string]string, error) {
	res := make(map[string]string)
	pairs := strings.Split(data, ";")

	for _, pair := range pairs {
		if pair == "" {
			continue
		}
		kv := strings.Split(pair, ":")
		if len(kv) != 2 {
			fn.Debug(fmt.Sprintf("Invalid record %s", pair))
			continue
		}

		res[kv[0]] = kv[1]
	}

	return res, nil
}

func (u *updater) fetchReleaseInfo() (map[string]string, error) {

	if u.releaseInfo != nil {
		return u.releaseInfo, nil
	}

	// TODO: use rest api for it
	s, err := net.LookupTXT("klversion.kloudlite.io")
	if err != nil {
		return nil, fn.NewE(err, "Failed to fetch txt records")
	}

	if len(s) == 0 {
		return nil, fn.Errorf("No records found for url klversion.kloudlite.io")
	}

	m, err := parseTxtResp(s[0])
	if err != nil {
		return nil, fn.NewE(err)
	}

	u.releaseInfo = m
	return m, nil
}

func (u *updater) CheckForUpdates() (bool, error) {
	// fetch dns records of txt records of url facebook.com

	relInfo, err := u.fetchReleaseInfo()
	if err != nil {
		return false, err
	}

	vcode, ok := relInfo["version"]
	if !ok {
		return false, fn.Errorf("Failed to fetch release info")
	}

	if vcode != flags.Version {
		return true, nil
	}

	return false, nil
}

func (u *updater) GetUpdateMessage() (*string, error) {

	relInfo, err := u.fetchReleaseInfo()
	if err != nil {
		fn.PrintError(err)
		return nil, err
	}

	latestVersion, ok := relInfo["version"]
	if !ok {
		fn.PrintError(fn.Errorf("Failed to fetch release info"))
		return nil, err
	}

	currentVersion := flags.Version

	resp := ""
	resp += text.Red(fmt.Sprintf("\nA new version of %s is available (%s -> %s)", flags.CliName, currentVersion, latestVersion))

	resp += "\n"
	resp += "\n"
	resp += "To update, run the following command:"
	resp += "\n"
	resp += text.Green(fmt.Sprintf("  %s", text.Bold("kl update")))
	resp += "\n"

	return &resp, nil
}
