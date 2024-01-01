package server

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/kloudlite/kl/lib/util"
)

func SyncKubeConfig(accName, clustName *string) (*string, error) {
	name := ""
	var err error

	if accName == nil {
		name, err = CurrentClusterName()
		if err != nil {
			return nil, err
		}
	} else {
		name = *accName
	}

	config, err := getKubeConfig(accName, clustName)
	if err != nil {
		return nil, err
	}

	tmpDir := os.TempDir()
	tmpFile := path.Join(tmpDir, name)

	_, err = os.Stat(tmpFile)
	if err == nil {
		return &tmpFile, nil
	}

	if err := os.WriteFile(tmpFile, []byte(*config), 0644); err != nil {
		log.Fatal(err)
	}

	return &tmpFile, nil
}

func getKubeConfig(accName, clusterName *string) (*string, error) {
	cookie, err := getCookie()
	if err != nil {
		return nil, err
	}

	if accName == nil {
		_, err = util.CurrentAccountName()
		if err != nil {
			return nil, err
		}
	}

	var currentCluster string

	if clusterName == nil {
		currentCluster, err = CurrentClusterName()
		if err != nil {
			return nil, err
		}
	} else {
		currentCluster = *clusterName
	}

	respData, err := klFetch("cli_getKubeConfig", map[string]any{
		"name": currentCluster,
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
			return nil, fmt.Errorf("cluster %s is not ready", currentCluster)
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
