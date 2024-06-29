package k3d

import (
	cluster "github.com/k3d-io/k3d/v5/cmd/cluster"
)

func CreateCluster() {
	createClusterCmd := cluster.NewCmdClusterCreate()
	createClusterCmd.SetArgs([]string{"create", "mycluster"})
	createClusterCmd.Execute()
}
