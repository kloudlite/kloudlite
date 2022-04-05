package application

import (
	"fmt"
	"go.uber.org/fx"
	"kloudlite.io/apps/infra/internal/domain"
	"os"
	"os/exec"
)

type infraClient struct {
}

func (i infraClient) CreateKubernetes(action domain.SetupClusterAction) error {
	initCommand := exec.Command(
		"terraform",
		fmt.Sprintf("-chdir=./internal/application/tf/create-cluster/%v", action.Provider),
		"init",
	)
	initCommand.Stdout = os.Stdout
	initCommand.Stderr = os.Stderr
	initCommand.Run()
	applyCommand := exec.Command(
		"terraform",
		fmt.Sprintf("-chdir=./internal/application/tf/create-cluster/%v", action.Provider),
		"apply",
		"-auto-approve",
		fmt.Sprintf("-var=cluster-id=%v", action.ClusterID),
		fmt.Sprintf("-var=master-nodes-count=%v", action.MastersCount),
		fmt.Sprintf("-var=agent-nodes-count=%v", action.NodesCount-action.MastersCount),
	)
	applyCommand.Stdout = os.Stdout
	applyCommand.Stderr = os.Stderr
	return applyCommand.Run()
}

func (i infraClient) SetupCSI(clusterId string) error {
	//TODO implement me
	panic("implement me")
}

func (i infraClient) SetupOperator(clusterId string) error {
	//TODO implement me
	panic("implement me")
}

func (i infraClient) SetupMonitoring(clusterId string) error {
	//TODO implement me
	panic("implement me")
}

func (i infraClient) SetupIngress(clusterId string) error {
	//TODO implement me
	panic("implement me")
}

func (i infraClient) SetupWireguard(clusterId string) error {
	//TODO implement me
	panic("implement me")
}

func newInfraClient() domain.TF {
	return &infraClient{}
}

var Module = fx.Module("applicaiton", fx.Provide(newInfraClient), domain.Module)
