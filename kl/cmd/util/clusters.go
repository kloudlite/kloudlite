package util

import (
	"fmt"

	"github.com/kloudlite/kl/lib"
	"github.com/kloudlite/kl/lib/server"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/pkg/errors"
)

func SelectCluster(args []string) (*ResourceData, error) {
	clusterName := ""
	if len(args) >= 1 {
		clusterName = args[0]
	}
	clusters, err := server.GetClusters()
	if err != nil {
		if err.Error() == "noSelectedAccount" {
			_, err := SelectAccount([]string{})
			if err != nil {
				return nil, err
			}
			return SelectCluster([]string{})
		}
		return nil, err
	}

	if clusterName != "" {
		for _, a := range clusters {
			if a.Metadata.Name == clusterName {
				return &ResourceData{
					Name:        a.Metadata.Name,
					DisplayName: a.DisplayName,
				}, nil
			}
		}
		return nil, errors.New("you don't have access to this cluster")
	}

	selectedIndex, err := fuzzyfinder.Find(
		clusters,
		func(i int) string {
			return fmt.Sprintf("%s %s", clusters[i].DisplayName, func() string {
				if clusters[i].Status.IsReady {
					return ""
				} else {
					return "(Not Ready)"
				}
			}())
		},
		fuzzyfinder.WithPromptString("Select Cluster > "),
	)

	if err != nil {
		return nil, err
	}

	if err = lib.SelectCluster(clusters[selectedIndex].Metadata.Name); err != nil {
		return nil, err
	}
	return &ResourceData{
		Name:        clusters[selectedIndex].Metadata.Name,
		DisplayName: clusters[selectedIndex].DisplayName,
	}, nil
}
