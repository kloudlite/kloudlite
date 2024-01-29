package client

import (
	"errors"
)

func CurrentClusterName() (string, error) {

	mc, err := GetMainCtx()
	if err != nil {
		return "", err
	}

	if mc.ClusterName == "" {
		return "", errors.New("please select a cluster using \"kl use cluster\"")
	}

	return mc.ClusterName, nil
}
