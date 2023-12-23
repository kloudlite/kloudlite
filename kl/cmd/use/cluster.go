/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package use

import (
	"errors"

	"github.com/kloudlite/kl/lib"
	"github.com/kloudlite/kl/lib/common"
	"github.com/kloudlite/kl/lib/server"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "select cluster to use later with all commands",
	Long: `Select cluster
Examples:
  # select cluster
  kl use cluster
	# this will open selector where you can select one of the cluster accessible to you.

  # select account with cluster id
  kl use cluster <clusterId>
	`,
	Run: func(_ *cobra.Command, args []string) {
		clusterName, err := SelectCluster(args)

		if err != nil {
			common.PrintError(err)
			return
		}

		err = lib.SelectCluster(clusterName)
		if err != nil {
			common.PrintError(err)
			return
		}

	},
}

func SelectCluster(args []string) (string, error) {
	clusterName := ""
	if len(args) >= 1 {
		clusterName = args[0]
	}
	clusters, err := server.GetClusters()
	if err != nil {
		return "", err
	}

	if clusterName != "" {
		for _, a := range clusters {
			if a.Metadata.Name == clusterName {
				return a.Metadata.Name, nil
			}
		}
		return "", errors.New("you don't have access to this cluster")
	}

	selectedIndex, err := fuzzyfinder.Find(
		clusters,
		func(i int) string {
			return clusters[i].DisplayName
		},
		fuzzyfinder.WithPromptString("Use Cluster >"),
	)

	if err != nil {
		return "", err
	}

	return clusters[selectedIndex].Metadata.Name, nil
}
