package application

import (
	"fmt"
	"go.uber.org/fx"
	"kloudlite.io/apps/infra/internal/domain"
	"kloudlite.io/pkg/config"
	"kloudlite.io/pkg/errors"
	"os"
	"os/exec"
)

type infraClient struct {
	env *InfraEnv
}

func (i *infraClient) CreateKubernetes(action domain.SetupClusterAction) (e error) {
	defer errors.HandleErr(&e)
	copyTemplateDirCommand := exec.Command(
		"cp",
		"-r",
		fmt.Sprintf("./internal/application/tf/%v", action.Provider),
		fmt.Sprintf("%v/%v", i.env.DataPath, action.ClusterID),
	)
	copyTemplateDirCommand.Stdout = os.Stdout
	copyTemplateDirCommand.Stderr = os.Stderr
	e = copyTemplateDirCommand.Run()
	errors.AssertNoError(e, fmt.Errorf("unable to copy template directory"))
	initCommand := exec.Command(
		"terraform",
		fmt.Sprintf("-chdir=%v/%v", i.env.DataPath, action.ClusterID),
		"init",
	)
	initCommand.Stdout = os.Stdout
	initCommand.Stderr = os.Stderr
	e = initCommand.Run()
	errors.AssertNoError(e, fmt.Errorf("failed to init terraform check"))
	applyCommand := exec.Command(
		"terraform",
		fmt.Sprintf("-chdir=%v/%v", i.env.DataPath, action.ClusterID),
		"apply",
		"-auto-approve",
		fmt.Sprintf("-var=cluster-id=%v", action.ClusterID),
		fmt.Sprintf("-var=master-nodes-count=%v", action.MastersCount),
		fmt.Sprintf("-var=agent-nodes-count=%v", action.NodesCount-action.MastersCount),
		fmt.Sprintf("-var=keys-path=%v", i.env.SshKeysPath),
		fmt.Sprintf("-var=do-token=%v", i.env.DoAPIKey),
	)
	applyCommand.Stdout = os.Stdout
	applyCommand.Stderr = os.Stderr
	e = applyCommand.Run()
	errors.AssertNoError(e, fmt.Errorf("failed to apply terraform"))
	e = i.CreateKubeConfig(action.ClusterID)
	errors.AssertNoError(e, fmt.Errorf("unable to create kubeconfig"))
	return e
}

func (i *infraClient) CreateKubeConfig(clusterId string) error {
	genKubeConfigCommand := exec.Command("terraform", "output", "-raw", "kube_config")
	outfile, err := os.Create(fmt.Sprintf("%v/%v/kube.yaml", i.env.DataPath, clusterId))
	if err != nil {
		return err
	}
	defer outfile.Close()
	genKubeConfigCommand.Stdout = outfile
	return genKubeConfigCommand.Run()
}

func (i *infraClient) UpdateKubernetes(action domain.UpdateClusterAction) (e error) {
	defer errors.HandleErr(&e)
	applyCommand := exec.Command(
		"terraform",
		fmt.Sprintf("-chdir=%v", fmt.Sprintf("%v/%v", i.env.DataPath, action.ClusterID)),
		"apply",
		"-auto-approve",
		fmt.Sprintf("-var=cluster-id=%v", action.ClusterID),
		fmt.Sprintf("-var=master-nodes-count=%v", action.MastersCount),
		fmt.Sprintf("-var=agent-nodes-count=%v", action.NodesCount-action.MastersCount),
		fmt.Sprintf("-var=keys-path=%v", i.env.SshKeysPath),
		fmt.Sprintf("-var=do-token=%v", i.env.DoAPIKey),
	)
	applyCommand.Stdout = os.Stdout
	applyCommand.Stderr = os.Stderr
	e = applyCommand.Run()
	errors.AssertNoError(e, fmt.Errorf("failed to apply terraform"))
	return e
}

func (i *infraClient) SetupCSI(clusterId string) error {
	//TODO implement me
	panic("implement me")
}

func (i *infraClient) SetupOperator(clusterId string) error {
	//TODO implement me
	panic("implement me")
}

func (i *infraClient) SetupMonitoring(clusterId string) error {
	//TODO implement me
	panic("implement me")
}

func (i *infraClient) SetupIngress(clusterId string) error {
	//TODO implement me
	panic("implement me")
}

func (i *infraClient) SetupWireguard(clusterId string) error {
	//TODO implement me
	panic("implement me")
}

func newInfraClient(env *InfraEnv) domain.TF {
	return &infraClient{
		env,
	}
}

type InfraEnv struct {
	DoAPIKey    string `env:"DO_API_KEY", required:"true"`
	DataPath    string `env:"DATA_PATH", required:"true"`
	SshKeysPath string `env:"SSH_KEYS_PATH", required:"true"`
}

var Module = fx.Module("applicaiton",
	// Load Env
	fx.Provide(func() (*InfraEnv, error) {
		var envC InfraEnv
		err := config.LoadConfigFromEnv(&envC)
		return &envC, err
	}),
	fx.Provide(newInfraClient),
	domain.Module,
)
