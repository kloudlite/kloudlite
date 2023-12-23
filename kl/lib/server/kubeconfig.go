package server

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
)

func SyncKubeConfig() (*string, error) {
	name, err := CurrentClusterName()

	if err != nil {
		return nil, err
	}
	tmpDir := os.TempDir()

	tmpFile := fmt.Sprintf("%s%s.config", tmpDir, name)

	config, err := getKubeConfig()
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(tmpFile, []byte(*config), 0644); err != nil {
		log.Fatal(err)
	}

	return &tmpFile, nil
}

func getKubeConfig() (*string, error) {
	cookie, err := getCookie()
	if err != nil {
		return nil, err
	}
	currentCluster, err := CurrentClusterName()
	if err != nil {
		return nil, err
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
	}
	if fromResp, err := GetFromResp[KubeConfig](respData); err != nil {
		return nil, err
	} else {
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
