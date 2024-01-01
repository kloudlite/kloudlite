package server

import (
	"errors"

	"github.com/kloudlite/kl/lib/util"
)

type Check struct {
	Generation int    `json:"generation"`
	Message    string `json:"message"`
}

type Cluster struct {
	Metadata struct {
		Name string `json:"name"`
	}
	DisplayName string `json:"displayName"`
	Status      struct {
		IsReady bool `json:"isReady"`
	} `json:"status"`
}

func CurrentClusterName() (string, error) {
	file, err := util.GetContextFile()
	if err != nil {
		return "", err
	}
	if file.ClusterName == "" {
		return "", errors.New("noSelectedCluster")
	}
	if file.ClusterName == "" {
		return "",
			errors.New("no accounts is selected yet. please select one using \"kl use account\"")
	}
	return file.ClusterName, nil
}

func GetClusters() ([]Cluster, error) {
	if _, err := util.CurrentAccountName(); err != nil {
		return nil, err
	}
	cookie, err := getCookie()
	if err != nil {
		return nil, err
	}
	respData, err := klFetch("cli_listClusters", map[string]any{
		"query": map[string]any{
			"first": 100,
		},
	}, &cookie)

	if err != nil {
		return nil, err
	}

	type ClusterList struct {
		Edges []struct {
			Node Cluster `json:"node"`
		} `json:"edges"`
	}
	if fromResp, err := GetFromResp[ClusterList](respData); err != nil {
		return nil, err
	} else {

		clusters := make([]Cluster, 0)
		for _, edge := range fromResp.Edges {
			clusters = append(clusters, edge.Node)
		}
		return clusters, nil
	}
}
