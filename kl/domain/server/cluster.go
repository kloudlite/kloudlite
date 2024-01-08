package server

import (
	"errors"
	"fmt"

	"github.com/kloudlite/kl/domain/client"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/fzf"
	"github.com/kloudlite/kl/pkg/ui/text"
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

func ListClusters() ([]Cluster, error) {
	s, _ := client.CurrentAccountName()
	if s == "" {
		if _, err := SelectAccount(s); err != nil {
			return nil, err
		}
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

func SelectCluster(clusterName string) (*Cluster, error) {
	clusters, err := ListClusters()

	if clusterName != "" {
		for _, a := range clusters {
			if a.Metadata.Name == clusterName {
				return &a, nil
			}
		}
		return nil, errors.New("you don't have access to this cluster")
	}

	c, err := fzf.FindOne(clusters,
		func(item Cluster) string {
			return fmt.Sprintf("%s (%s) %s",
				item.DisplayName, item.Metadata.Name,

				func() string {
					if !item.Status.IsReady {
						return "not ready to use"
					}
					return ""
				}(),
			)
		},
		fzf.WithPrompt(text.Green("Select Cluster > ")),
	)
	if err != nil {
		return nil, err
	}

	if err := client.SelectCluster(c.Metadata.Name); err != nil {
		return nil, err
	}

	return c, nil
}

func EnsureCluster(options ...fn.Option) (string, error) {
	accountName := fn.GetOption(options, "accountName")

	_, err := EnsureAccount(accountName)
	if err != nil {
		return "", err
	}

	clusterName := fn.GetOption(options, "clusterName")

	if clusterName != "" {
		return clusterName, nil
	}

	clusterName, _ = client.CurrentClusterName()

	if clusterName != "" {
		return clusterName, nil
	}

	c, err := SelectCluster(clusterName)

	if err != nil {
		return "", err
	}

	return c.Metadata.Name, nil
}
