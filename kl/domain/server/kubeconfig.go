package server

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path"

	fn "github.com/kloudlite/kl/pkg/functions"
)

func SyncKubeConfig(options ...fn.Option) (*string, error) {

	accountName := fn.GetOption(options, "accountName")
	clusterName := fn.GetOption(options, "clusterName")

	accountName, err := EnsureAccount(accountName)
	if err != nil {
		return nil, err
	}

	clusterName, err = EnsureCluster(options...)
	if err != nil {
		return nil, err
	}

	tmpDir := os.TempDir()
	tmpFile := path.Join(tmpDir, clusterName)

	_, err = os.Stat(tmpFile)
	if err == nil {
		return &tmpFile, nil
	}

	config, err := getKubeConfig(options...)
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(tmpFile, []byte(*config), 0644); err != nil {
		log.Fatal(err)
	}

	return &tmpFile, nil
}

func getKubeConfig(options ...fn.Option) (*string, error) {

	accountName := fn.GetOption(options, "accountName")
	clusterName := fn.GetOption(options, "clusterName")

	_, err := EnsureAccount(accountName)
	if err != nil {
		return nil, err
	}

	clusterName, err = EnsureCluster(options...)
	if err != nil {
		return nil, err
	}

	cookie, err := getCookie()
	if err != nil {
		return nil, err
	}

	respData, err := klFetch("cli_getKubeConfig", map[string]any{
		"name": clusterName,
	}, &cookie)
	if err != nil {
		return nil, err
	}

	type KubeConfig struct {
		AdminKubeConfig struct {
			Encoding string `json:"encoding"`
			Value    string `json:"value"`
		} `json:"adminKubeConfig"`
		Status struct {
			IsReady bool `json:"isReady"`
		} `json:"status"`
	}
	if fromResp, err := GetFromResp[KubeConfig](respData); err != nil {
		return nil, err
	} else {

		if !(*fromResp).Status.IsReady {
			return nil, fmt.Errorf("cluster %s is not ready", clusterName)
		}

		value := (*fromResp).AdminKubeConfig.Value
		encoding := (*fromResp).AdminKubeConfig.Encoding
		switch encoding {
		case "base64":
			if value, err = base64Decode(value); err != nil {
				return nil, err
			} else {
				return &value, nil
			}
		default:
			return nil, fmt.Errorf("unknown encoding %s", encoding)
		}
	}
}

func base64Decode(in string) (string, error) {
	decodeString, err := base64.StdEncoding.DecodeString(in)
	if err != nil {
		return "", err
	}
	return string(decodeString), nil
}
