package common_cmd

import (
	"fmt"
	"github.com/kloudlite/kl/domain/server"
	"github.com/kloudlite/kl/pkg/ui/fzf"
	"github.com/kloudlite/kl/pkg/ui/text"

	"github.com/kloudlite/kl/lib"
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

	c, err := fzf.FindOne(clusters,
		func(item server.Cluster) string {
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

	if err != nil {
		return nil, err
	}

	if err = lib.SelectCluster(c.Metadata.Name); err != nil {
		return nil, err
	}
	return &ResourceData{
		Name:        c.Metadata.Name,
		DisplayName: c.DisplayName,
	}, nil
}
